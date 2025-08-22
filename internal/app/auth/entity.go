package auth

import (
	"time"

	"github.com/google/uuid"
)

type authUserEntity struct {
	Username    string
	Password    string
	Permissions []string
}

type authSessionEntity struct {
	ID           uuid.UUID
	Username     string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RefreshToken string
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
