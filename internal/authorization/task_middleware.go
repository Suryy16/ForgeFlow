package authorization

import (
	"context"
	"errors"
	"net/http"

	"github.com/Suryy16/ForgeFlow/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskAccessStore interface {
	GetTaskProjectRole(
		ctx context.Context,
		taskID uuid.UUID,
		userID uuid.UUID,
	) (string, error)
}

type postgresTaskAccessStore struct {
	db *pgxpool.Pool
}

func NewPostgresTaskAccessStore(db *pgxpool.Pool) TaskAccessStore {
	return &postgresTaskAccessStore{db: db}
}

func (s *postgresTaskAccessStore) GetTaskProjectRole(
	ctx context.Context,
	taskID uuid.UUID,
	userID uuid.UUID,
) (string, error) {
	var role string

	query := `
		SELECT pm.role
		FROM tasks t
		JOIN project_members pm
			ON pm.project_id = t.project_id
		WHERE t.id = $1
		  AND pm.user_id = $2
	`

	err := s.db.QueryRow(ctx, query, taskID, userID).Scan(&role)

	if errors.Is(err, pgx.ErrNoRows) {
		// Task tidak ditemukan atau user bukan anggota project.
		return "", ErrTaskNotFound
	}

	if err != nil {
		return "", err
	}

	return role, nil
}

func RequireTaskRole(
	store TaskAccessStore,
	taskIDParam string,
	allowedRoles ...string,
) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedRoles))

	for _, role := range allowedRoles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			taskID, err := uuid.Parse(r.PathValue(taskIDParam))
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid task ID")
				return
			}

			role, err := store.GetTaskProjectRole(
				r.Context(),
				taskID,
				userID,
			)

			if errors.Is(err, ErrTaskNotFound) {
				writeError(w, http.StatusNotFound, "task not found")
				return
			}

			if err != nil {
				writeError(
					w,
					http.StatusInternalServerError,
					"failed to verify task access",
				)
				return
			}

			if !allowed[role] {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
