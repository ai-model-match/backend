package useCase

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errUseCaseSameCodeAlreadyExists = errors.New("use-case-same-code-already-exists")
var errUseCaseCodeChangeNotAllowedWhileActive = errors.New("use-case-code-change-not-allowed-while-active")
var errUseCaseCannotBeDeletedWhileActive = errors.New("use-case-cannot-be-deleted-while-active")
var errUseCaseCannotBeActivatedWithoutActiveFlow = errors.New("use-case-cannot-be-activated-without-active-flow")
