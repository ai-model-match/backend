package mm_pubsub

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

/*
PubSubEventType represents an event type that can be published or consumed within the pub-sub system.
Generally, the PubSubEventType is related to an event entity and the possible actions performed on it.
It is preferable to use the past participle to indicate that the event was generated as a result
of an application state change.
*/
type PubSubEventType string

/*
List of avaiable events can be published and consumed within the pub-sub system.
*/
const (
	UseCaseCreatedEvent         PubSubEventType = "use-case.created"
	UseCaseUpdatedEvent         PubSubEventType = "use-case.updated"
	UseCaseDeletedEvent         PubSubEventType = "use-case.deleted"
	UseCaseStepCreatedEvent     PubSubEventType = "use-case-step.created"
	UseCaseStepUpdatedEvent     PubSubEventType = "use-case-step.updated"
	UseCaseStepDeletedEvent     PubSubEventType = "use-case-step.deleted"
	FlowCreatedEvent            PubSubEventType = "flow.created"
	FlowUpdatedEvent            PubSubEventType = "flow.updated"
	FlowDeletedEvent            PubSubEventType = "flow.deleted"
	FlowStatisticsCreatedEvent  PubSubEventType = "flow-statistics.created"
	FlowStatisticsUpdatedEvent  PubSubEventType = "flow-statistics.updated"
	FlowStepCreatedEvent        PubSubEventType = "flow-step.created"
	FlowStepUpdatedEvent        PubSubEventType = "flow-step.updated"
	FlowStepDeletedEvent        PubSubEventType = "flow-step.deleted"
	RolloutStrategyCreatedEvent PubSubEventType = "rollout-strategy.created"
	RolloutStrategyUpdatedEvent PubSubEventType = "rollout-strategy.updated"
	PickerMatchedEvent          PubSubEventType = "picker.matched"
	FeedbackCreatedEvent        PubSubEventType = "feedback.created"
)

/*
Map each event type to a function that returns a pointer to the right struct.
It is useful for unmarshal stored events and replay
*/
var eventEntityFactories = map[PubSubEventType]func() any{
	UseCaseCreatedEvent:         func() interface{} { return &UseCaseEventEntity{} },
	UseCaseUpdatedEvent:         func() interface{} { return &UseCaseEventEntity{} },
	UseCaseDeletedEvent:         func() interface{} { return &UseCaseEventEntity{} },
	UseCaseStepCreatedEvent:     func() interface{} { return &UseCaseStepEventEntity{} },
	UseCaseStepUpdatedEvent:     func() interface{} { return &UseCaseStepEventEntity{} },
	UseCaseStepDeletedEvent:     func() interface{} { return &UseCaseStepEventEntity{} },
	FlowCreatedEvent:            func() interface{} { return &FlowEventEntity{} },
	FlowUpdatedEvent:            func() interface{} { return &FlowEventEntity{} },
	FlowDeletedEvent:            func() interface{} { return &FlowEventEntity{} },
	FlowStatisticsCreatedEvent:  func() interface{} { return &FlowStatisticsEventEntity{} },
	FlowStatisticsUpdatedEvent:  func() interface{} { return &FlowStatisticsEventEntity{} },
	FlowStepCreatedEvent:        func() interface{} { return &FlowStepEventEntity{} },
	FlowStepUpdatedEvent:        func() interface{} { return &FlowStepEventEntity{} },
	FlowStepDeletedEvent:        func() interface{} { return &FlowStepEventEntity{} },
	RolloutStrategyCreatedEvent: func() interface{} { return &RolloutStrategyEventEntity{} },
	RolloutStrategyUpdatedEvent: func() interface{} { return &RolloutStrategyEventEntity{} },
	PickerMatchedEvent:          func() interface{} { return &PickerEventEntity{} },
	FeedbackCreatedEvent:        func() interface{} { return &FeedbackEventEntity{} },
}

/*
PubSubEvent represents a generic struct for events. All the events must be structured in this way,
ensuring the payload of the event itself is stored inside the EventEntity.
*/
type PubSubEvent struct {
	EventID     uuid.UUID       `json:"eventId"`
	EventTime   time.Time       `json:"eventTime"`
	EventType   PubSubEventType `json:"eventType"`
	EventEntity interface{}     `json:"eventEntity"`
	EventState  *sync.WaitGroup `json:"-"`
}
