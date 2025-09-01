package picker

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errUseCaseNotAcive = errors.New("use-case-not-active")
var errUseCaseStepNotFound = errors.New("use-case-step-not-found")
var errFlowNotFound = errors.New("flow-not-found")
var errCorrelationConflict = errors.New("correlation-conflict")
var errFlowsNotAvailable = errors.New("flows-not-available")
var errFallbackFlowNotAvailable = errors.New("fallback-flow-not-available")
