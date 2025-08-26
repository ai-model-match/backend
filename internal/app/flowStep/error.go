package flowStep

import "errors"

var errFlowNotFound = errors.New("flow-not-found")
var errFlowStepNotFound = errors.New("flow-step-not-found")
var errFlowStepWrongConfigFormat = errors.New("flow-step-wrong-config-format")
