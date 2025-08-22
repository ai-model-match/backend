package mm_auth

import (
	"strings"

	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var authJwtSecret string

/*
Initialize the AuthMiddleware by setting the JWT secret needed to validate the received token
*/
func InitAuthMiddleware(secret string) {
	authJwtSecret = secret
}

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
AuthMiddleware Middleware on APIs to check if the user is authenticated
and verify the permissions the user has compared to the permissions required
by the API. In case of failure, returns an error to the client.
This guard verifies if the user can access the Company identified by its ID
available in the API url.
*/
func AuthMiddleware(permissionsToCheck []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve the authenticated user
		authenticatedUser, err := getAuthenticatedUserFromRequest(ctx)
		// In case of error or if the user is not found, return Unauthorized
		if err != nil || mm_utils.IsEmpty(authenticatedUser) {
			mm_router.ReturnUnauthorizedError(ctx)
			return
		}
		// If permissions are not defined, return Forbidden by default.
		if len(permissionsToCheck) == 0 {
			mm_router.ReturnForbiddenError(ctx)
			return
		}
		// Check if all the required permissions are included in the authenticated User permissions, otherwise return Forbidden
		if !containsAll(authenticatedUser.Permissions, permissionsToCheck) {
			mm_router.ReturnForbiddenError(ctx)
			return
		}
		ctx.Set(contextAuthenticatedUser, &authenticatedUser)
		ctx.Next()
	}
}

/*
Retrieve the authenticated user from the request.
*/
func getAuthenticatedUserFromRequest(ctx *gin.Context) (AuthenticatedUser, error) {
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
		return []byte(authJwtSecret), nil
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

/*
GetAuthenticatedUserFromSession retrieves the authenticated user from the session.
This works in combination of the Authentication middleware that extracts all the information
provided by the JWT sent in the Authentication header of the request and store them
in the request context. This utility retrieve the authenticated user from the context session
without performing additional read operations to get all the users information.
*/
func GetAuthenticatedUserFromSession(ctx *gin.Context) *AuthenticatedUser {
	value, exists := ctx.Get(contextAuthenticatedUser)
	if exists {
		return value.(*AuthenticatedUser)
	}
	return nil
}
