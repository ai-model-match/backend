package mm_auth

/*
contextAuthenticatedUser represents a key where the authenticated user information
are stored inside the context of the request.
*/
const contextAuthenticatedUser = "authenticatedUser"

/*
AuthenticatedUser represents an authenticated user in the webapp application.
All the information stored here are retrieved by the
JWT in the Authentication header of the request.
*/
type AuthenticatedUser struct {
	Username    string
	Permissions []string
}

/*
List of permissions we can leverage to evaluate if an authenticated user can perform a specific operation
before performing the API logic.
*/
const (
	READ      = "read"
	WRITE     = "write"
	REFRESH   = "refresh"
	M2M_READ  = "m2m_read"
	M2M_WRITE = "m2m_write"
)
