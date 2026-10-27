package project

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
	repo ProjectRepository
}

func NewHandler(repo ProjectRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := r.PathValue("id")
	id, err := ParseUUID(rawID)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		http.Error(
			w, "Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("failed to encode project response: %v", err)
	}
}

func (h *Handler) GetProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projects, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		log.Printf("failed to encode projects response: %v", err)
	}
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateProjectRequest

	decoder := json.NewDecoder(
		http.MaxBytesReader(w, r.Body, 1<<20),
	)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain exactly one JSON object",
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Name is required",
		})
		return
	}

	if len(req.Name) > 255 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Name must not exceed 255 characters",
		})
		return
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		req.Description = &description
	}

	var ownerID *uuid.UUID

	if req.OwnerID != nil {
		value := strings.TrimSpace(*req.OwnerID)
		if value == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "owner_id must be a valid UUID or null",
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

		ownerID = &parsedID
	}

	project := Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}

	created, err := h.repo.Create(r.Context(), project)

	if err != nil {
		switch {
		case errors.Is(err, ErrOwnerNotMember):
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "Owner is not member",
			})
		default:
			fmt.Printf("CreateProject failed: %v\n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create project",
			})
		}
		return
	}

	location := fmt.Sprintf(
		"/api/v1/projects/%s",
		created.ID.String(),
	)
	w.Header().Set("Location", location)

	writeJSON(w, http.StatusCreated, toProjectResponse(created))
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

func (h *Handler) UpdateProject(
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
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var input UpdateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if input.Name == "" {
		http.Error(
			w,
			"Name is required",
			http.StatusBadRequest,
		)
		return
	}

	project := Project{
		ID:          id,
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Description: input.Description,
	}

	updated, err := h.repo.Update(r.Context(), project)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		log.Printf("UpdateProject failed: %v", err)

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

func (h *Handler) DeleteProject(
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
			"Invalid project ID",
			http.StatusBadRequest,
		)
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			http.Error(
				w,
				"Project not found",
				http.StatusNotFound,
			)
			return
		}

		log.Printf("failed to delete project %s: %v", id, err)

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
