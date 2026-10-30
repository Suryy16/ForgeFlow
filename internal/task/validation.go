package task

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidTitle      = errors.New("title is required or too long")
	ErrInvalidPriority   = errors.New("invalid priority")
	ErrInvalidAssigneeID = errors.New("invalid assignee ID")
	errInvalidUUID       = errors.New("Invalid UUID")
)

func ParseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, errInvalidUUID
	}
	return id, nil
}

func ValidateCreateTaskInput(req *CreateTaskRequest) (*uuid.UUID, error) {
	req.Title = strings.TrimSpace(req.Title)

	if req.Title == "" || len(req.Title) > 255 {
		return nil, ErrInvalidTitle
	}

	req.Priority = strings.ToUpper(strings.TrimSpace(req.Priority))

	if req.Priority == "" {
		req.Priority = "MEDIUM"
	}

	switch req.Priority {
	case "LOW", "MEDIUM", "HIGH", "URGENT":
	default:
		return nil, ErrInvalidPriority
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		req.Description = &description
	}

	if req.AssigneeID == nil {
		return nil, nil
	}

	value := strings.TrimSpace(*req.AssigneeID)
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, ErrInvalidAssigneeID
	}

	return &id, nil
}
