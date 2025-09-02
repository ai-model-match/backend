package mm_pubsub

const (
	RolloutStateInit            RolloutState = "INIT"
	RolloutStateWarmup          RolloutState = "WARMUP"
	RolloutStateEscaped         RolloutState = "ESCAPED"
	RolloutStateAdaptive        RolloutState = "ADAPTIVE"
	RolloutStateCompleted       RolloutState = "COMPLETED"
	RolloutStateForcedEscaped   RolloutState = "FORCED_ESCAPED"
	RolloutStateForcedCompleted RolloutState = "FORCED_COMPLETED"
)

var AvailableRolloutState = []interface{}{
	RolloutStateInit,
	RolloutStateWarmup,
	RolloutStateEscaped,
	RolloutStateAdaptive,
	RolloutStateCompleted,
	RolloutStateForcedEscaped,
	RolloutStateForcedCompleted,
}
