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
	TopicUseCaseV1     PubSubTopic = "topic/v1/use-case"
	TopicUseCaseStepV1 PubSubTopic = "topic/v1/use-case-step"
)
