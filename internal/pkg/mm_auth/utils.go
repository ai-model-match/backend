package mm_auth

import (
	"strings"

	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

/*
Given two array of string, check if all the elements of the subset
are contained in the entire set
*/
func containsAll(set []string, subset []string) bool {
	m := map[string]struct{}{}
	for _, v := range set {
		m[v] = struct{}{}
	}
	for _, v := range subset {
		if _, ok := m[v]; !ok {
			return false
		}
	}
	return true
}

/*
Retrieve the authenticated user from the request.
*/
func getAuthenticatedUserFromRequest(ctx *gin.Context) (AuthenticatedUser, error) {
	// Check Bearer token
	user, err := getAuthenticatedUserFromJWTAuth(ctx)
	if mm_utils.IsEmpty(user) {
		// If auth user is not fund, fallback on API Key
		return getAuthenticatedUserFromApiKey(ctx)
	}
	return user, err
}

func getAuthenticatedUserFromJWTAuth(ctx *gin.Context) (AuthenticatedUser, error) {
	// Extract and check the Authorization format (begins with "Bearer")
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return AuthenticatedUser{}, nil
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return AuthenticatedUser{}, nil
	}

	// Parse the token and validate it with the private key
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return AuthenticatedUser{}, nil
		}
		return []byte(authConfig.JwtSecret), nil
	})
	// if the token is not valid, return
	if err != nil || !token.Valid {
		return AuthenticatedUser{}, nil
	}

	// Extract the Username from claims
	username, _ := claims["sub"].(string)

	// Extract permissions from claims
	var permissions []string
	if c, ok := claims["permissions"].([]interface{}); ok {
		for _, v := range c {
			if s, ok := v.(string); ok {
				permissions = append(permissions, s)
			}
		}
	}
	return AuthenticatedUser{
		Username:    username,
		Permissions: permissions,
	}, nil
}

func getAuthenticatedUserFromApiKey(ctx *gin.Context) (AuthenticatedUser, error) {
	// Extract and check the Authorization format (begins with "ApiKey")
	tokenString := ctx.GetHeader("X-Api-Key")
	if tokenString == "" {
		return AuthenticatedUser{}, nil
	}
	switch tokenString {
	case authConfig.ApiKeyReadOnly:
		return AuthenticatedUser{
			Username:    authConfig.ApiKeyReadOnlyUsername,
			Permissions: []string{M2M_READ},
		}, nil
	case authConfig.ApiKeyReadOnly:
		return AuthenticatedUser{
			Username:    authConfig.ApiKeyReadWriteUsername,
			Permissions: []string{M2M_READ, M2M_WRITE},
		}, nil
	default:
		return AuthenticatedUser{}, nil
	}
}
