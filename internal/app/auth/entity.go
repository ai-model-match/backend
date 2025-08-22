package auth

import (
	"time"

	"github.com/google/uuid"
)

// Available claims to inject in JWT for Auth User
type Claim string

const READ Claim = "read"
const WRITE Claim = "write"
const REFRESH Claim = "refresh"

type authUserEntity struct {
	Username string
	Password string
	Claims   []Claim
}

type authTokenEntity struct {
	AccessToken           string    `json:"accessToken"`
	RefreshToken          string    `json:"refreshToken"`
	AccessTokenID         uuid.UUID `json:"-"`
	RefreshTokenID        uuid.UUID `json:"-"`
	AccessTokenCreatedAt  time.Time `json:"-"`
	RefreshTokenCreatedAt time.Time `json:"-"`
	AccessTokenExpiresAt  time.Time `json:"-"`
	RefreshTokenExpiresAt time.Time `json:"-"`
}

type authEntity struct {
	ID           uuid.UUID
	Username     string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RefreshToken string
}
