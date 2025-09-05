package rolloutStrategy

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errRolloutStrategyNotFound = errors.New("rollout-strategy-not-found")
var errRolloutStrategyAlreadyExists = errors.New("rollout-strategy-already-exists")
var errRolloutStrategyNotEditableWhileActive = errors.New("rollout-strategy-not-editable-while-active")
var errRolloutStrategyTransitionStateNotAllowed = errors.New("rollout-strategy-transition-state-not-allowed")
