package flowStatistics

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type flowStatisticsConsumerInterface interface {
	subscribe()
}

type flowStatisticsConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service flowStatisticsServiceInterface
}

func newFlowStatisticsConsumer(pubSub *mm_pubsub.PubSubAgent, service flowStatisticsServiceInterface) flowStatisticsConsumer {
	consumer := flowStatisticsConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r flowStatisticsConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicFlowV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.L().Error("Panic occured in handling a new message", zap.String("service", "flow-statistics-consumer"))
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "flow-statistics-consumer"),
					)
					return
				}
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "flow-statistics-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.FlowCreatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.FlowEventEntity)
				// Create the Flow Statistics
				if _, err := r.service.createFlowStatistics(event.ID); err != nil {
					if err != errFlowStatisticsAlreadyExists {
						zap.L().Error("Impossible to create the flowStatisticss for the new Flow", zap.String("service", "flow-statistics-consumer"), zap.Error(err))
						return
					}
				}
			}()
		}
	}()
}
