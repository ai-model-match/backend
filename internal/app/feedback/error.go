package feedback

import "errors"

var errCorrelationNotFound = errors.New("correlation-not-found")
var errFeedbackAlreadyProvided = errors.New("feedback-already-provided")
