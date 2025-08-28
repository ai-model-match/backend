package flow

import (
	"time"

	"github.com/google/uuid"
)

type flowEntity struct {
	ID              uuid.UUID  `json:"id"`
	UseCaseID       uuid.UUID  `json:"useCaseId"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Active          *bool      `json:"active"`
	Fallback        *bool      `json:"fallback"`
	InitialServePct *float64   `json:"initialServePct"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	ClonedFromID    *uuid.UUID `json:"-"`
}
