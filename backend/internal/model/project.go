package model

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	CreatedAt   time.Time  `json:"created_at"`
	Tasks       []Task     `json:"tasks,omitempty"` // populated only on GET /projects/:id
}
