package flowStep

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type flowStepConsumerInterface interface {
	subscribe()
}

type flowStepConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service flowStepServiceInterface
}

func newFlowStepConsumer(pubSub *mm_pubsub.PubSubAgent, service flowStepServiceInterface) flowStepConsumer {
	consumer := flowStepConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r flowStepConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicUseCaseStepV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						mm_log.LogPanicError(r, "flow-step-consumer", "Panic occurred in handling a new message")
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "flow-step-consumer"),
					)
					return
				}
				// ACK message
				defer msg.Message.EventState.Done()
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "flow-step-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.UseCaseStepCreatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.UseCaseStepEventEntity)
				// Create any missing FLow step compared to Use Case steps
				if err := r.service.createStepsForAllFlowsOfUseCase(event.UseCaseID); err != nil {
					zap.L().Error("Impossible to create all flowSteps for the new Use Case Step", zap.String("service", "flow-step-consumer"))
					return
				}
			}()
		}
	}()

	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicFlowV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						mm_log.LogPanicError(r, "flow-step-consumer", "Panic occurred in handling a new message")
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "flow-step-consumer"),
					)
					return
				}
				// ACK message
				defer msg.Message.EventState.Done()
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "flow-step-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.FlowCreatedEvent {
					return
				}

				event := msg.Message.EventEntity.(*mm_pubsub.FlowEventEntity)
				// If the created event is a cloned one, clone all its steps
				if event.ClonedFromID != nil {
					if err := r.service.cloneStepsFromFlow(event.ID, *event.ClonedFromID); err != nil {
						zap.L().Error("Impossible to clone all flowSteps for the new cloned Flow", zap.String("service", "flow-step-consumer"))
						return
					}
				}
				// Create any missing FLow step compared to Use Case steps
				if err := r.service.createStepsForAllFlowsOfUseCase(event.UseCaseID); err != nil {
					zap.L().Error("Impossible to create all flowSteps for the new Flow", zap.String("service", "flow-step-consumer"))
					return
				}
			}()
		}
	}()
}
