package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type rolloutStrategyConsumerInterface interface {
	subscribe()
}

type rolloutStrategyConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service rolloutStrategyServiceInterface
}

func newRolloutStrategyConsumer(pubSub *mm_pubsub.PubSubAgent, service rolloutStrategyServiceInterface) rolloutStrategyConsumer {
	consumer := rolloutStrategyConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r rolloutStrategyConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicUseCaseV1)
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
				if msg.Message.EventType != mm_pubsub.UseCaseCreatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.UseCaseEventEntity)
				// Create the Rollout Strategy
				if _, err := r.service.createRolloutStrategy(event.ID); err != nil {
					if err == errRolloutStrategyAlreadyExists {
						zap.L().Info("rolloutStrategy already exists. Skip event", zap.String("service", "rollout-strategy-consumer"))
						return
					} else {
						zap.L().Error("Impossible to create the rolloutStrategy for the new Use Case", zap.String("service", "rollout-strategy-consumer"))
						return
					}
				}
			}()
		}
	}()

	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicRsEngineV1)
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
				if err := r.service.updateRolloutStrategyFromEvent(*event); err != nil {
					zap.L().Error("Impossible to update the rolloutStrategy from RS Engine event", zap.String("service", "rollout-strategy-consumer"))
					return
				}
			}()
		}
	}()
}
