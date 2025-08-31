package mm_pubsub

const (
	RolloutStateInit            RolloutState = "INIT"
	RolloutStateWarmup          RolloutState = "WARMUP"
	RolloutStateEscaped         RolloutState = "ESCAPED"
	RolloutStateMonitor         RolloutState = "MONITOR"
	RolloutStateAdaptive        RolloutState = "ADAPTIVE"
	RolloutStateCompleted       RolloutState = "COMPLETED"
	RolloutStateForcedEscaped   RolloutState = "FORCED_ESCAPED"
	RolloutStateForcedCompleted RolloutState = "FORCED_COMPLETED"
)

var AvailableRolloutState = []interface{}{
	RolloutStateInit,
	RolloutStateWarmup,
	RolloutStateEscaped,
	RolloutStateMonitor,
	RolloutStateAdaptive,
	RolloutStateCompleted,
	RolloutStateForcedEscaped,
	RolloutStateForcedCompleted,
}
