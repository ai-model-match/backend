package rolloutStrategy

import "errors"

var errUseCaseNotFound = errors.New("use-case-not-found")
var errRolloutStrategyNotFound = errors.New("rollout-strategy-not-found")
var errRolloutStrategyWrongConfigFormat = errors.New("rollout-strategy-wrong-config-format")
var errRolloutStrategyAlreadyExists = errors.New("rollout-strategy-already-exists")
