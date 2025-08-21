package mmerr

import "errors"

/*
ErrGeneric represents a generic error across the entire application.
*/
var ErrGeneric = errors.New("generic-error")
