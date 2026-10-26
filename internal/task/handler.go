package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Handler struct {
	repo TaskRepository
}

func NewHandler(repo TaskRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("failed to encode tasks response: %v", err)
	}
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := r.PathValue("id")
	id, err := ParseUUID(rawID)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("failed to encode task response: %v", err)
	}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// 1. Parse project_id dari URL.
	projectID, err := uuid.Parse(r.PathValue("project_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "project_id must be a valid UUID",
		})
		return
	}

	// 2. Decode request body.
	var req CreateTaskRequest

	decoder := json.NewDecoder(
		http.MaxBytesReader(w, r.Body, 1<<20), // Maksimum 1 MB
	)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Tolak lebih dari satu JSON object dalam body.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain exactly one JSON object",
		})
		return
	}

	// 3. Validasi dan normalisasi input.
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "title is required",
		})
		return
	}

	if len(req.Title) > 255 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "title must not exceed 255 characters",
		})
		return
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		req.Description = &description
	}

	req.Priority = strings.ToUpper(strings.TrimSpace(req.Priority))
	if req.Priority == "" {
		req.Priority = "MEDIUM"
	}

	switch req.Priority {
	case "LOW", "MEDIUM", "HIGH", "URGENT":
		// Valid.
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "priority must be LOW, MEDIUM, HIGH, or URGENT",
		})
		return
	}

	// 4. Parse assignee_id jika diberikan.
	var assigneeID *uuid.UUID

	if req.AssigneeID != nil {
		value := strings.TrimSpace(*req.AssigneeID)
		if value == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "assignee_id must be a valid UUID or null",
			})
			return
		}

		parsedID, err := uuid.Parse(value)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "assignee_id must be a valid UUID",
			})
			return
		}

		assigneeID = &parsedID
	}

	// 5. Bentuk model Task.
	task := Task{
		ProjectID:   projectID,
		AssigneeID:  assigneeID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "TODO",
		Priority:    req.Priority,
	}

	// 6. Simpan task melalui repository.
	created, err := h.repo.Create(r.Context(), task)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "project not found",
			})
		case errors.Is(err, ErrAssigneeNotMember):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "assignee must be a member of the project",
			})
		default:
			// Detail error disimpan di log server, bukan dikirim ke client.
			fmt.Printf("CreateTask failed: %v\n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create task",
			})
		}
		return
	}

	// 7. Kirim response 201 Created.
	location := fmt.Sprintf(
		"/api/v1/tasks/%s",
		created.ID.String(),
	)
	w.Header().Set("Location", location)

	writeJSON(w, http.StatusCreated, toTaskResponse(created))
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Status sudah dikirim, jadi jangan mencoba mengirim response
		// HTTP kedua. Kegagalan ini sebaiknya dicatat di logger.
		fmt.Printf("write JSON response failed: %v\n", err)
	}
}

func (h *Handler) UpdateTask(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPut {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	rawID := r.PathValue("id")
	id, err := ParseUUID(rawID)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var input UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	// Validasi dasar.
	if input.Title == "" {
		http.Error(
			w,
			"Title is required",
			http.StatusBadRequest,
		)
		return
	}

	task := Task{
		ID:          id,
		ProjectID:   input.ProjectID,
		AssigneeID:  input.AssigneeID,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		Priority:    input.Priority,
		DueDate:     input.DueDate,
	}

	updated, err := h.repo.Update(r.Context(), task)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		log.Printf("UpdateTask failed: %v", err)

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		log.Printf("failed to encode updated task: %v", err)
	}
}

func (h *Handler) DeleteTask(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodDelete {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	rawID := r.PathValue("id")
	id, err := ParseUUID(rawID)
	if err != nil {
		http.Error(
			w,
			"Invalid task ID",
			http.StatusBadRequest,
		)
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(
				w,
				"Task not found",
				http.StatusNotFound,
			)
			return
		}
		log.Printf("failed to delete task %s: %v", id, err)

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
