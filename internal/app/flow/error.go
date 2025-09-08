package flow

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errFlowNotFound = errors.New("flow-not-found")
var errActiveFlowNotFound = errors.New("active-flow-not-found")
var errFlowCannotBeDeletedIfActive = errors.New("flow-cannot-be-deleted-if-active")
var errFlowCannotBeDeactivatedIfLastActive = errors.New("flow-cannot-be-deactivated-if-last-active")
