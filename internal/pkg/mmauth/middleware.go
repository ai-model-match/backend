package mmauth

import (
	"github.com/ai-model-match/backend/internal/pkg/mmrouter"
	"github.com/ai-model-match/backend/internal/pkg/mmutils"
	"github.com/gin-gonic/gin"
)

/*
AuthMiddleware Middleware on APIs to check if the user is authenticated
and verify the permissions the user has compared to the permissions required
by the API. In case of failure, returns an error to the client.
This guard verifies if the user can access the Company identified by its ID
available in the API url.

@TODO The logic of checking if the user exists and its claim needs to be
implemented based on the auth service you will use.
*/
func AuthMiddleware(claimsToCheck []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Retrieve the authenticated user
		authUser, err := getAuthUserFromRequest(ctx)
		// In case of error or if the user is not found, return Unauthorized
		if err != nil || mmutils.IsEmpty(authUser) {
			mmrouter.ReturnUnauthorizedError(ctx)
			return
		}
		// If claims are not valid or missing, return Forbidden.
		if len(claimsToCheck) == 0 {
			mmrouter.ReturnForbiddenError(ctx)
			return
		}
		ctx.Set(contextAuthUser, authUser)
		ctx.Next()
	}
}

/*
Retrieve the authenticated user from the request.
Here you can implement the logic to retrieve the user info from the JWT
and additional information from the database (or from external auth service like supertokens, auth0, etc.).

@TODO This logic needs to be implemented based on the Auth system you will use.
*/
func getAuthUserFromRequest(_ *gin.Context) (AuthUser, error) {
	return AuthUser{}, nil
}

/*
GetAuthUserFromSession retrieves the authenticated user from the session.
This works in combination of the Authentication middleware that extracts all the information
provided by the JWT sent in the Authentication header of the request and store them
in the request context. This utility retrieve the authenticated user from the context session
without performing additional read operations to get all the users information.
*/
func GetAuthUserFromSession(ctx *gin.Context) *AuthUser {
	value, exists := ctx.Get(contextAuthUser)
	if exists {
		return value.(*AuthUser)
	}
	return nil
}
