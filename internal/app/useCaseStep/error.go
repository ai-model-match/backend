package useCaseStep

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errUseCaseStepNotFound = errors.New("use-case-step-not-found")
var errUseCaseStepSameCodeAlreadyExists = errors.New("use-case-step-same-code-already-exists")
