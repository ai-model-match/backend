package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type authUtilInterface interface {
	generateToken(user authUserEntity) (authTokenEntity, error)
}

type authUtil struct {
	authJwtSecret               string
	authJwtAccessTokenDuration  int
	authJwtRefreshTokenDuration int
}

func newAuthUtil(authJwtSecret string, authJwtAccessTokenDuration int, authJwtRefreshTokenDuration int) authUtil {
	return authUtil{
		authJwtSecret:               authJwtSecret,
		authJwtAccessTokenDuration:  authJwtAccessTokenDuration,
		authJwtRefreshTokenDuration: authJwtRefreshTokenDuration,
	}
}

func (u authUtil) generateToken(user authUserEntity) (authTokenEntity, error) {
	// Define JWT Claims including permissions
	type CustomClaims struct {
		jwt.RegisteredClaims
		Permissions []Claim `json:"permissions"`
	}

	// Define Access and Refresh Token ID and their duration
	now := time.Now()
	accessTokenExpiresAt := now.Add(time.Duration(u.authJwtAccessTokenDuration) * time.Second)
	refreshTokenExpiresAt := now.Add(time.Duration(u.authJwtRefreshTokenDuration) * time.Second)
	accessTokenID := uuid.New()
	refreshTokenID := uuid.New()

	// Create Access Token with claims
	accessClaims := CustomClaims{
		Permissions: user.Claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenID.String(),
			Issuer:    "ai-model-match",
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(u.authJwtSecret))
	if err != nil {
		return authTokenEntity{}, err
	}

	// Create Refresh Token with claims
	refreshClaims := CustomClaims{
		Permissions: []Claim{REFRESH},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenID.String(),
			Issuer:    "ai-model-match",
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(u.authJwtSecret))
	if err != nil {
		return authTokenEntity{}, err
	}

	// Return generated tokens
	return authTokenEntity{
		AccessToken:           accessStr,
		RefreshToken:          refreshStr,
		AccessTokenID:         accessTokenID,
		RefreshTokenID:        refreshTokenID,
		AccessTokenCreatedAt:  now,
		RefreshTokenCreatedAt: now,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}, nil
}
