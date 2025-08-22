package mm_pubsub

import (
	"time"

	"github.com/google/uuid"
)

/*
UserEventEntity represents a User entity in pub-sub system.
*/
type UseCaseEventEntity struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
