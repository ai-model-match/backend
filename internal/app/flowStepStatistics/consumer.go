package flowStepStatistics

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type flowStepStatisticsConsumerInterface interface {
	subscribe()
}

type flowStepStatisticsConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service flowStepStatisticsServiceInterface
}

func newFlowStepStatisticsConsumer(pubSub *mm_pubsub.PubSubAgent, service flowStepStatisticsServiceInterface) flowStepStatisticsConsumer {
	consumer := flowStepStatisticsConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r flowStepStatisticsConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicFlowStepV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.L().Error("Panic occured in handling a new message", zap.String("service", "flow-step-statistics-consumer"))
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "flow-step-statistics-consumer"),
					)
					return
				}
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "flow-step-statistics-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.FlowStepCreatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.FlowStepEventEntity)
				// Create the Flow Step Statistics
				if _, err := r.service.createFlowStepStatistics(event.ID); err != nil {
					if err == errFlowStepStatisticsAlreadyExists {
						zap.L().Info("flowStepStatistics already exists. Skip event", zap.String("service", "flow-step-statistics-consumer"))
						return
					} else {
						zap.L().Error("Impossible to create the flowStepStatistics for the new Flow Step", zap.String("service", "flow-step-statistics-consumer"))
						return
					}
				}
			}()
		}
	}()

	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicPickerV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.L().Error("Panic occured in handling a new message", zap.String("service", "flow-step-statistics-consumer"))
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "flow-step-statistics-consumer"),
					)
					return
				}
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "flow-step-statistics-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.PickerMatchedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.PickerEventEntity)
				// Create the Flow Step Statistics
				if err := r.service.updateStatistics(*event); err != nil {
					zap.L().Error("Impossible to update Flow Step statistics", zap.String("service", "flow-step-statistics-consumer"))
					return
				}
			}()
		}
	}()
}
