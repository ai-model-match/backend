package user

import (
	"github.com/ai-model-match/backend/internal/pkg/mmpubsub"
	"github.com/ai-model-match/backend/internal/pkg/mmutils"
	"go.uber.org/zap"
)

type userConsumerInterface interface {
	subscribe()
}

type userConsumer struct {
	pubSub  *mmpubsub.PubSubAgent
	service userServiceInterface
}

func newUserConsumer(pubSub *mmpubsub.PubSubAgent, service userServiceInterface) userConsumer {
	consumer := userConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r userConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(mmpubsub.TopicUserV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.L().Error("Panic occured in handling a new message", zap.String("service", "user-consumer"))
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "user-consumer"),
					)
					return
				}
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "user-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != mmpubsub.UserCreatedEvent {
					return
				}

				event := msg.Message.EventEntity.(mmpubsub.UserEventEntity)
				userID := event.ID
				input := createUserInputDto{
					ID:        mmutils.GetStringFromUUID(event.ID),
					Firstname: event.Firstname,
					Lastname:  event.Lastname,
					Email:     event.Email,
				}
				_, err := r.service.createUser(msg.Context, userID, input)
				if err != nil {
					zap.L().Error("Impossible to create a new user", zap.String("service", "user-consumer"))
					return
				}
			}()
		}
	}()
}
