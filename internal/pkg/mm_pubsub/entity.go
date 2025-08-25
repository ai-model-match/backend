package mm_pubsub

import (
	"time"

	"github.com/google/uuid"
)

type UseCaseEventEntity struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UseCaseStepEventEntity struct {
	ID          uuid.UUID `json:"id"`
	UseCaseID   uuid.UUID `json:"useCaseId"`
	Title       string    `json:"title"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Position    int64     `json:"position"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
