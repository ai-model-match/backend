package auth

import "errors"

var errExpiredRefreshToken = errors.New("expired-refresh-token")
var errInvalidCredentials = errors.New("invalid-username-or-password")
