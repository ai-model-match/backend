package mmauth

import (
	"time"

	"github.com/google/uuid"
)

/*
contextAuthUser represents a key where the authenticated user information
are stored inside the context of the request.
*/
const contextAuthUser = "authUser"

/*
AuthUser represents an authenticated user in the webapp application.
All the information stored here are retrieved by the
JWT in the Authentication header of the request.
*/
type AuthUser struct {
	ID        uuid.UUID
	Email     string
	Firstname string
	Lastname  string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

/*
List of claims we can leverage to evaluate if an authenticated user can perform a specific operation
before performing the API logic.
*/
const (
	UserGet    = "user-g"
	UserUpdate = "user-u"
	UserDelete = "user-d"
)
