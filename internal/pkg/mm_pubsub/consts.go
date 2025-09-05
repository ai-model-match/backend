package mm_pubsub

const (
	RolloutStateInit            RolloutState = "INIT"
	RolloutStateWarmup          RolloutState = "WARMUP"
	RolloutStateEscaped         RolloutState = "ESCAPED"
	RolloutStateAdaptive        RolloutState = "ADAPTIVE"
	RolloutStateCompleted       RolloutState = "COMPLETED"
	RolloutStateForcedStop      RolloutState = "FORCED_STOP"
	RolloutStateForcedEscaped   RolloutState = "FORCED_ESCAPED"
	RolloutStateForcedCompleted RolloutState = "FORCED_COMPLETED"
)

var AvailableRolloutState = []interface{}{
	RolloutStateInit,
	RolloutStateWarmup,
	RolloutStateEscaped,
	RolloutStateAdaptive,
	RolloutStateCompleted,
	RolloutStateForcedStop,
	RolloutStateForcedEscaped,
	RolloutStateForcedCompleted,
}

var AllowedTransitions = map[RolloutState][]RolloutState{
	RolloutStateInit:            {RolloutStateWarmup},
	RolloutStateWarmup:          {RolloutStateForcedStop, RolloutStateForcedEscaped, RolloutStateForcedCompleted},
	RolloutStateAdaptive:        {RolloutStateForcedStop, RolloutStateForcedEscaped, RolloutStateForcedCompleted},
	RolloutStateEscaped:         {RolloutStateInit},
	RolloutStateCompleted:       {RolloutStateInit},
	RolloutStateForcedStop:      {RolloutStateInit},
	RolloutStateForcedEscaped:   {RolloutStateInit},
	RolloutStateForcedCompleted: {RolloutStateInit},
}
