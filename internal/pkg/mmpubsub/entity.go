package mmpubsub

import (
	"time"

	"github.com/google/uuid"
)

/*
UserEventEntity represents a User entity in pub-sub system.
*/
type UserEventEntity struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	CreatedBy uuid.UUID  `json:"createdBy"`
	UpdatedBy uuid.UUID  `json:"updatedBy"`
	DeletedBy *uuid.UUID `json:"deletedBy"`
}
