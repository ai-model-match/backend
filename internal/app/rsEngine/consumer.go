package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_log"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type rsEngineConsumerInterface interface {
	subscribe()
}

type rsEngineConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service rsEngineServiceInterface
}

func newRsEngineConsumer(pubSub *mm_pubsub.PubSubAgent, service rsEngineServiceInterface) rsEngineConsumer {
	consumer := rsEngineConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r rsEngineConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicFlowStatisticsV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						mm_log.LogPanicError(r, "rs-engine-consumer", "Panic occurred in handling a new message")
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "rs-engine-consumer"),
					)
					return
				}
				// ACK message
				defer msg.Message.EventState.Done()
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "rs-engine-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.FlowStatisticsUpdatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.FlowStatisticsEventEntity)
				// Create the Rollout Strategy
				if err := r.service.onFlowStatisticsUpdate(*event); err != nil {
					zap.L().Error("Impossible to run the rsEngine for the new updated statistics", zap.String("service", "rs-engine-consumer"))
					return
				}
			}()
		}
	}()

	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicRolloutStrategyV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						mm_log.LogPanicError(r, "rs-engine-consumer", "Panic occurred in handling a new message")
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "rs-engine-consumer"),
					)
					return
				}
				// ACK message
				defer msg.Message.EventState.Done()
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "rs-engine-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.RolloutStrategyUpdatedEvent {
					return
				}
				event := msg.Message.EventEntity.(*mm_pubsub.RolloutStrategyEventEntity)
				if event.RolloutState != mm_pubsub.RolloutStateForcedEscaped {
					zap.L().Info("Ignoring Rollout Strategy Event. Only Forced Escaped events are processed", zap.String("service", "rs-engine-consumer"))
					return
				}
				// Create the Rollout Strategy
				if err := r.service.onRolloutStrategyForcedEscaped(*event); err != nil {
					zap.L().Error("Impossible to run the rsEngine for the new updated statistics", zap.String("service", "rs-engine-consumer"))
					return
				}
			}()
		}
	}()
}
