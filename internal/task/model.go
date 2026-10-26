package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	AssigneeID  *uuid.UUID
	Title       string
	Description *string
	Status      string
	Priority    string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func toTaskResponse(t Task) TaskResponse {
	var assigneeID *string
	if t.AssigneeID != nil {
		value := t.AssigneeID.String()
		assigneeID = &value
	}

	var dueDate *string
	if t.DueDate != nil {
		value := t.DueDate.Format("2006-01-02")
		dueDate = &value
	}

	return TaskResponse{
		ID:          t.ID.String(),
		ProjectID:   t.ProjectID.String(),
		AssigneeID:  assigneeID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		DueDate:     dueDate,
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
