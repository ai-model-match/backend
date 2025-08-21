package user

import (
	"time"

	"github.com/google/uuid"
)

type userEntity struct {
	id        uuid.UUID
	email     string
	firstname string
	lastname  string
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
	createdBy uuid.UUID
	updatedBy uuid.UUID
	deletedBy *uuid.UUID
}
