package useCaseStep

import (
	"math"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_err"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type useCaseStepServiceInterface interface {
	listUseCaseSteps(ctx *gin.Context, input ListUseCaseStepsInputDto) ([]useCaseStepEntity, int64, error)
	getUseCaseStepByID(ctx *gin.Context, input getUseCaseStepInputDto) (useCaseStepEntity, error)
	createUseCaseStep(ctx *gin.Context, input createUseCaseStepInputDto) (useCaseStepEntity, error)
	updateUseCaseStep(ctx *gin.Context, input updateUseCaseStepInputDto) (useCaseStepEntity, error)
	deleteUseCaseStep(ctx *gin.Context, input deleteUseCaseStepInputDto) (useCaseStepEntity, error)
}

type useCaseStepService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  useCaseStepRepositoryInterface
}

func newUseCaseStepService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository useCaseStepRepositoryInterface) useCaseStepService {
	return useCaseStepService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s useCaseStepService) listUseCaseSteps(ctx *gin.Context, input ListUseCaseStepsInputDto) ([]useCaseStepEntity, int64, error) {
	useCaseID := uuid.MustParse(input.UseCaseID)
	exists, err := s.repository.checkUseCaseExists(s.storage, useCaseID)
	if err != nil {
		return []useCaseStepEntity{}, 0, mm_err.ErrGeneric
	}
	if !exists {
		return []useCaseStepEntity{}, 0, errUseCaseNotFound
	}
	limit, offset := mm_utils.PagePageSizeToLimitOffset(input.Page, input.PageSize)
	items, totalCount, err := s.repository.listUseCaseSteps(s.storage, useCaseID, limit, offset, useCaseStepOrderBy(input.OrderBy), mm_db.OrderDir(input.OrderDir), input.SearchKey, false)
	if err != nil || items == nil {
		return []useCaseStepEntity{}, 0, mm_err.ErrGeneric
	}
	return items, totalCount, nil
}

func (s useCaseStepService) getUseCaseStepByID(ctx *gin.Context, input getUseCaseStepInputDto) (useCaseStepEntity, error) {
	useCaseStepID := uuid.MustParse(input.ID)
	item, err := s.repository.getUseCaseStepByID(s.storage, useCaseStepID, false)
	if err != nil {
		return useCaseStepEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return useCaseStepEntity{}, errUseCaseStepNotFound
	}
	return item, nil
}

func (s useCaseStepService) createUseCaseStep(ctx *gin.Context, input createUseCaseStepInputDto) (useCaseStepEntity, error) {
	now := time.Now()
	useCaseID := uuid.MustParse(input.UseCaseID)
	useCaseStep := useCaseStepEntity{
		ID:          uuid.New(),
		UseCaseID:   useCaseID,
		Title:       input.Title,
		Code:        input.Code,
		Description: input.Description,
		Position:    math.MaxInt64,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		exists, err := s.repository.checkUseCaseExists(s.storage, useCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !exists {
			return errUseCaseNotFound
		}
		item, err := s.repository.getUseCaseStepByCode(tx, useCaseID, input.Code, false)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(item) {
			return errUseCaseStepSameCodeAlreadyExists
		}
		if _, err = s.repository.saveUseCaseStep(tx, useCaseStep); err != nil {
			return mm_err.ErrGeneric
		}
		if err := s.repository.recalculateUseCaseStepPosition(tx, useCaseID); err != nil {
			return mm_err.ErrGeneric
		}
		if useCaseStep, err = s.repository.getUseCaseStepByID(tx, useCaseStep.ID, false); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseStepEntity{}, errTransaction
	}
	go s.pubSubAgent.Publish(mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mm_pubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mm_pubsub.UseCaseStepCreatedEvent,
			EventEntity: mm_pubsub.UseCaseStepEventEntity{
				ID:          useCaseStep.ID,
				UseCaseID:   useCaseStep.UseCaseID,
				Title:       useCaseStep.Title,
				Code:        useCaseStep.Code,
				Description: useCaseStep.Description,
				Position:    useCaseStep.Position,
				CreatedAt:   useCaseStep.CreatedAt,
				UpdatedAt:   useCaseStep.UpdatedAt,
			},
		},
	})
	return useCaseStep, nil
}

func (s useCaseStepService) updateUseCaseStep(ctx *gin.Context, input updateUseCaseStepInputDto) (useCaseStepEntity, error) {
	now := time.Now()
	var useCaseStep useCaseStepEntity
	err_transaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case Step exists
		useCaseStepID := uuid.MustParse(input.ID)
		item, err := s.repository.getUseCaseStepByID(tx, useCaseStepID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errUseCaseStepNotFound
		}
		// If the input contains a new code for the use case, check for collision
		if input.Code != nil {
			useCaseStepSameCode, err := s.repository.getUseCaseStepByCode(tx, item.UseCaseID, *input.Code, false)
			if err != nil {
				return mm_err.ErrGeneric
			}
			if !mm_utils.IsEmpty(useCaseStepSameCode) && useCaseStepSameCode.ID.String() != input.ID {
				return errUseCaseStepSameCodeAlreadyExists
			}
		}
		// Update useCaseStep information based on inputs
		useCaseStep = item
		useCaseStep.UpdatedAt = now
		if input.Title != nil {
			useCaseStep.Title = *input.Title
		}
		if input.Description != nil {
			useCaseStep.Description = *input.Description
		}
		if input.Code != nil {
			useCaseStep.Code = *input.Code
		}
		if input.Position != nil {
			// If the step is moving in a lower position (e.g. from 10 to 3),
			// we need to move it one step more, so that, the algorith to re-sort all steps correctly
			if useCaseStep.Position > *input.Position {
				*input.Position = *input.Position - 1
			}
			useCaseStep.Position = *input.Position
		}
		if _, err = s.repository.saveUseCaseStep(tx, useCaseStep); err != nil {
			return mm_err.ErrGeneric
		}
		if err := s.repository.recalculateUseCaseStepPosition(tx, useCaseStep.UseCaseID); err != nil {
			return mm_err.ErrGeneric
		}
		if useCaseStep, err = s.repository.getUseCaseStepByID(tx, useCaseStep.ID, false); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if err_transaction != nil {
		return useCaseStepEntity{}, err_transaction
	}
	// Send an event of useCaseStep updated
	go s.pubSubAgent.Publish(mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mm_pubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mm_pubsub.UseCaseStepUpdatedEvent,
			EventEntity: mm_pubsub.UseCaseStepEventEntity{
				ID:          useCaseStep.ID,
				UseCaseID:   useCaseStep.UseCaseID,
				Title:       useCaseStep.Title,
				Code:        useCaseStep.Code,
				Description: useCaseStep.Description,
				Position:    useCaseStep.Position,
				CreatedAt:   useCaseStep.CreatedAt,
				UpdatedAt:   useCaseStep.UpdatedAt,
			},
		},
	})
	return useCaseStep, nil
}

func (s useCaseStepService) deleteUseCaseStep(ctx *gin.Context, input deleteUseCaseStepInputDto) (useCaseStepEntity, error) {
	var useCaseStep useCaseStepEntity
	err_transaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseStepID := uuid.MustParse(input.ID)
		item, err := s.repository.getUseCaseStepByID(tx, useCaseStepID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errUseCaseStepNotFound
		}
		useCaseStep = item
		if _, err := s.repository.deleteUseCaseStep(tx, useCaseStep); err != nil {
			return mm_err.ErrGeneric
		}
		if err := s.repository.recalculateUseCaseStepPosition(tx, item.UseCaseID); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if err_transaction != nil {
		return useCaseStepEntity{}, err_transaction
	}
	// Send an event of useCaseStep updated
	go s.pubSubAgent.Publish(mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mm_pubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mm_pubsub.UseCaseStepDeletedEvent,
			EventEntity: mm_pubsub.UseCaseStepEventEntity{
				ID:          useCaseStep.ID,
				UseCaseID:   useCaseStep.UseCaseID,
				Title:       useCaseStep.Title,
				Code:        useCaseStep.Code,
				Description: useCaseStep.Description,
				Position:    useCaseStep.Position,
				CreatedAt:   useCaseStep.CreatedAt,
				UpdatedAt:   useCaseStep.UpdatedAt,
			},
		},
	})
	return useCaseStep, nil
}
