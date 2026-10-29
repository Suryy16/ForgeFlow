package project

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMember struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
