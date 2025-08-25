package flow

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errFlowNotFound = errors.New("flow-not-found")
var errFlowCannotDeleteIfFallbackAndUseCaseActive = errors.New("flow-cannot-delete-if-fallback-and-use-case-active")
var errFlowCannotRemoveFallbackWithActiveUseCase = errors.New("flow-cannot-remove-fallback-with-active-use-case")
