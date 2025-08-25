package useCase

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
)

type useCaseConsumerInterface interface {
	subscribe()
}

type useCaseConsumer struct {
	pubSub  *mm_pubsub.PubSubAgent
	service useCaseServiceInterface
}

func newUseCaseConsumer(pubSub *mm_pubsub.PubSubAgent, service useCaseServiceInterface) useCaseConsumer {
	consumer := useCaseConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r useCaseConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mm_pubsub.TopicUseCaseV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.L().Error("Panic occured in handling a new message", zap.String("service", "use-case-consumer"))
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "use-case-consumer"),
					)
					return
				}
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "use-case-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mm_pubsub.UseCaseCreatedEvent {
					return
				}

				/*event := msg.Message.EventEntity.(mm_pubsub.UseCaseEventEntity)
				useCaseID := event.ID
				input := createUseCaseInputDto{
					ID:          mm_utils.GetStringFromUUID(event.ID),
					Title:       event.Title,
					Code:        event.Code,
					Description: event.Description,
				}
				_, err := r.service.createUseCase(msg.Context, useCaseID, input)
				if err != nil {
					zap.L().Error("Impossible to create a new useCase", zap.String("service", "use-case-consumer"))
					return
				}*/
			}()
		}
	}()
}
