package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Handler struct {
	repo ProjectRepository
}

func NewHandler(repo ProjectRepository) *Handler {
	return &Handler{repo: repo}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func requireMethod(
	w http.ResponseWriter,
	r *http.Request,
	method string,
) bool {
	if r.Method == method {
		return true
	}

	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
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
	if !requireMethod(w, r, http.MethodGet) {
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

func (h *Handler) CreateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req CreateProjectRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(
			w,
			http.StatusBadRequest,
			"request body must contain exactly one JSON object",
		)
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if utf8.RuneCountInString(req.Name) > 150 {
		writeError(
			w,
			http.StatusBadRequest,
			"name must not exceed 150 characters",
		)
		return
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		req.Description = &description
	}

	if req.OwnerID == nil || strings.TrimSpace(*req.OwnerID) == "" {
		writeError(w, http.StatusBadRequest, "owner_id is required")
		return
	}

	ownerID, err := uuid.Parse(strings.TrimSpace(*req.OwnerID))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"owner_id must be a valid UUID",
		)
		return
	}

	project := Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     &ownerID,
	}

	created, err := h.repo.Create(r.Context(), project)
	if err != nil {
		if errors.Is(err, ErrOwnerNotMember) {
			writeError(
				w,
				http.StatusConflict,
				"project owner must be a project member",
			)
			return
		}

		log.Printf("CreateProject failed: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"failed to create project",
		)
		return
	}

	w.Header().Set(
		"Location",
		fmt.Sprintf("/api/v1/projects/%s", created.ID),
	)

	writeJSON(w, http.StatusCreated, toProjectResponse(created))
}

func (h *Handler) UpdateProject(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !requireMethod(w, r, http.MethodPut) {
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
	if !requireMethod(w, r, http.MethodDelete) {
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

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	projectID, err := uuid.Parse(r.PathValue("project_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	var req AddProjectMemberRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(
			w,
			http.StatusBadRequest,
			"request body must contain exactly one JSON object",
		)
		return
	}

	userID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		writeError(w, http.StatusBadRequest, "user_id must be a valid UUID")
		return
	}

	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if role == "" {
		role = "MEMBER"
	}

	if role != "ADMIN" && role != "MEMBER" {
		writeError(w, http.StatusBadRequest, "role must be ADMIN or MEMBER")
		return
	}

	member := ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	created, err := h.repo.AddMember(r.Context(), member)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, ErrMemberExists):
			writeError(
				w,
				http.StatusConflict,
				"user is already a member of this project",
			)
		default:
			log.Printf("AddMember failed: %v", err)
			writeError(
				w,
				http.StatusInternalServerError,
				"failed to add project member",
			)
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"project_id": created.ProjectID,
		"user_id":    created.UserID,
		"role":       created.Role,
		"created_at": created.CreatedAt,
		"updated_at": created.UpdatedAt,
	})
}
