package flow

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type flowConsumerInterface interface {
	subscribe()
}

type flowConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service flowServiceInterface
}

func newFlowConsumer(pubSub *mm_pubsub.PubSubAgent, service flowServiceInterface) flowConsumer {
	consumer := flowConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r flowConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicRsEnginekV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						mm_log.LogPanicError(r, "rollout-strategy-consumer", "Panic occurred in handling a new message")
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "rollout-strategy-consumer"),
					)
					return
				}
				// ACK message
				defer msg.Message.EventState.Done()
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "rollout-strategy-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.RsEngineUpdatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.RsEngineEventEntity)
				// Update the Rollout Strategy
				if err := r.service.updateFlowsFromEvent(*event); err != nil {
					zap.L().Error("Impossible to update flows from RS Engine event", zap.String("service", "rollout-strategy-consumer"))
					return
				}
			}()
		}
	}()
}
