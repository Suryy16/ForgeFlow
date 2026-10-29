package authorization

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Suryy16/ForgeFlow/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectRoleStore reads a user's role within a specific project.
type ProjectRoleStore interface {
	GetProjectRole(ctx context.Context, projectID, userID uuid.UUID) (string, error)
}

type postgresProjectRoleStore struct {
	db *pgxpool.Pool
}

func NewPostgresProjectRoleStore(db *pgxpool.Pool) ProjectRoleStore {
	return &postgresProjectRoleStore{db: db}
}

func (s *postgresProjectRoleStore) GetProjectRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	var role string
	err := s.db.QueryRow(ctx, `
		SELECT role
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotProjectMember
	}
	return role, err
}

var ErrNotProjectMember = errors.New("user is not a member of this project")

// RequireProjectRole verifies that the authenticated user has one of the allowed
// roles in the project identified by the given Go 1.22 ServeMux path parameter.
// Authentication middleware must run before this middleware.
func RequireProjectRole(store ProjectRoleStore, projectIDParam string, allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[strings.ToUpper(strings.TrimSpace(role))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			projectID, err := uuid.Parse(r.PathValue(projectIDParam))
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid project ID")
				return
			}

			role, err := store.GetProjectRole(r.Context(), projectID, userID)
			if errors.Is(err, ErrNotProjectMember) {
				writeError(w, http.StatusForbidden, "you are not a member of this project")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to verify project permissions")
				return
			}

			if _, ok := allowed[strings.ToUpper(role)]; !ok {
				writeError(w, http.StatusForbidden, "insufficient project role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = writeJSONBody(w, map[string]string{"error": message})
}
