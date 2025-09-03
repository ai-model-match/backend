package mm_pubsub

/*
PubSubTopic represents a topic name where events are published and
consumed by different modules. Each topic must contain only events
related to a specific entity domain.
*/
type PubSubTopic string

/*
List of available topics.
*/
const (
	TopicUseCaseV1         PubSubTopic = "topic/v1/use-case"
	TopicUseCaseStepV1     PubSubTopic = "topic/v1/use-case-step"
	TopicFlowV1            PubSubTopic = "topic/v1/flow"
	TopicFlowStatisticsV1  PubSubTopic = "topic/v1/flow-statistics"
	TopicFlowStepV1        PubSubTopic = "topic/v1/flow-step"
	TopicRolloutStrategyV1 PubSubTopic = "topic/v1/rollout-strategy"
	TopicPickerV1          PubSubTopic = "topic/v1/picker"
	TopicFeedbackV1        PubSubTopic = "topic/v1/feedback"
)
