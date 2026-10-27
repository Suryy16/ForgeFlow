package project

import "github.com/google/uuid"

type CreateProjectRequest struct {
	OwnerID     *string `json:"owner_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type ProjectResponse struct {
	ID          string  `json:"id"`
	OwnerID     *string `json:"owner_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type UpdateProjectInput struct {
	OwnerID     *uuid.UUID `json:"owner_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
}
