package user

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmerr"
	"github.com/ai-model-match/backend/internal/pkg/mmpubsub"
	"github.com/ai-model-match/backend/internal/pkg/mmutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userServiceInterface interface {
	getUserByID(ctx *gin.Context, input getUserInputDto) (userEntity, error)
	createUser(ctx *gin.Context, requesterID uuid.UUID, input createUserInputDto) (userEntity, error)
}

type userService struct {
	storage     *gorm.DB
	pubSubAgent *mmpubsub.PubSubAgent
	repository  userRepositoryInterface
}

func newUserService(storage *gorm.DB, pubSubAgent *mmpubsub.PubSubAgent, repository userRepositoryInterface) userService {
	return userService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s userService) getUserByID(ctx *gin.Context, input getUserInputDto) (userEntity, error) {
	userID := uuid.MustParse(input.ID)
	item, err := s.repository.getUserByID(s.storage, userID, false)
	if err != nil {
		return userEntity{}, mmerr.ErrGeneric
	}
	if mmutils.IsEmpty(item) {
		return userEntity{}, errUserNotFound
	}
	return item, nil
}

func (s userService) createUser(ctx *gin.Context, requesterID uuid.UUID, input createUserInputDto) (userEntity, error) {
	now := time.Now()
	user := userEntity{
		ID:        uuid.MustParse(input.ID),
		Firstname: input.Firstname,
		Lastname:  input.Lastname,
		Email:     input.Email,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
		CreatedBy: requesterID,
		UpdatedBy: requesterID,
		DeletedBy: nil,
	}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		_, err := s.repository.saveUser(tx, user)
		if err != nil {
			return mmerr.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return userEntity{}, errTransaction
	}
	go s.pubSubAgent.Publish(mmpubsub.TopicUserV1, mmpubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mmpubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mmpubsub.UserCreatedEvent,
			EventEntity: mmpubsub.UserEventEntity{
				ID:        user.ID,
				Firstname: user.Firstname,
				Lastname:  user.Lastname,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				DeletedAt: user.DeletedAt,
				CreatedBy: user.CreatedBy,
				UpdatedBy: user.UpdatedBy,
				DeletedBy: user.DeletedBy,
			},
		},
	})
	return user, nil
}
