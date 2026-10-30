package task

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fakeTaskRepository struct {
	createFn       func(context.Context, Task) (Task, error)
	getByIDFn      func(context.Context, uuid.UUID) (Task, error)
	getAllFn       func(context.Context) ([]Task, error)
	GetAllByUserFn func(ctx context.Context, userID uuid.UUID) ([]Task, error)
	updateFn       func(context.Context, Task) (Task, error)
	deleteFn       func(context.Context, uuid.UUID) error
}

func (f *fakeTaskRepository) Create(
	ctx context.Context, task Task,
) (Task, error) {
	return f.createFn(ctx, task)
}

func (f *fakeTaskRepository) GetByID(
	ctx context.Context, id uuid.UUID,
) (Task, error) {
	return f.getByIDFn(ctx, id)
}

func (f *fakeTaskRepository) GetAll(
	ctx context.Context,
) ([]Task, error) {
	return f.getAllFn(ctx)
}

func (f *fakeTaskRepository) GetAllByUser(
	ctx context.Context, id uuid.UUID,
) ([]Task, error) {
	return f.GetAllByUserFn(ctx, id)
}

func (f *fakeTaskRepository) Update(
	ctx context.Context, task Task,
) (Task, error) {
	return f.updateFn(ctx, task)
}

func (f *fakeTaskRepository) Delete(
	ctx context.Context, id uuid.UUID,
) error {
	return f.deleteFn(ctx, id)
}

func TestDeleteTask(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		deleteErr      error
		expectedStatus int
	}{
		{
			name:           "success",
			taskID:         taskID.String(),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "task not found",
			taskID:         taskID.String(),
			deleteErr:      ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repository error",
			taskID:         taskID.String(),
			deleteErr:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeTaskRepository{
				deleteFn: func(
					ctx context.Context,
					id uuid.UUID,
				) error {
					return tt.deleteErr
				},
			}

			handler := NewHandler(repo)

			req := httptest.NewRequest(
				http.MethodDelete,
				"/api/v1/tasks/"+tt.taskID,
				nil,
			)

			// Simulasi path parameter dari Go 1.22+ ServeMux.
			req.SetPathValue("id", tt.taskID)

			rec := httptest.NewRecorder()

			handler.DeleteTask(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}
		})
	}
}

func TestCreateTask_InvalidBody(t *testing.T) {
	repo := &fakeTaskRepository{
		createFn: func(
			ctx context.Context,
			task Task,
		) (Task, error) {
			t.Fatal("repository should not be called")
			return Task{}, nil
		},
	}

	handler := NewHandler(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/projects/"+uuid.NewString()+"/tasks",
		strings.NewReader(`{"title": ""}`),
	)

	req.SetPathValue("project_id", uuid.NewString())

	rec := httptest.NewRecorder()

	handler.CreateTask(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
