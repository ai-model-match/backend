package mm_err

import "errors"

/*
ErrGeneric represents a generic error across the entire application.
*/
var ErrGeneric = errors.New("generic-error")
var ErrGenericInput = errors.New("invalid input format")
