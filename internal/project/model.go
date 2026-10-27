package project

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	OwnerID     *uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func toProjectResponse(p Project) ProjectResponse {
	var ownerID *string
	if p.OwnerID != nil {
		value := p.OwnerID.String()
		ownerID = &value
	}

	return ProjectResponse{
		ID:          p.ID.String(),
		OwnerID:     ownerID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
