package useCase

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_err"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type useCaseServiceInterface interface {
	listUseCases(ctx *gin.Context, input ListUseCasesInputDto) ([]useCaseEntity, int64, error)
	getUseCaseByID(ctx *gin.Context, input getUseCaseInputDto) (useCaseEntity, error)
	createUseCase(ctx *gin.Context, input createUseCaseInputDto) (useCaseEntity, error)
	updateUseCase(ctx *gin.Context, input updateUseCaseInputDto) (useCaseEntity, error)
	deleteUseCase(ctx *gin.Context, input deleteUseCaseInputDto) (useCaseEntity, error)
}

type useCaseService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  useCaseRepositoryInterface
}

func newUseCaseService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository useCaseRepositoryInterface) useCaseService {
	return useCaseService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s useCaseService) listUseCases(ctx *gin.Context, input ListUseCasesInputDto) ([]useCaseEntity, int64, error) {
	limit, offset := mm_utils.PagePageSizeToLimitOffset(input.Page, input.PageSize)
	items, totalCount, err := s.repository.listUseCases(s.storage, limit, offset, useCaseOrderBy(input.OrderBy), mm_db.OrderDir(input.OrderDir), input.SearchKey, false)
	if err != nil || items == nil {
		return []useCaseEntity{}, 0, mm_err.ErrGeneric
	}
	return items, totalCount, nil
}

func (s useCaseService) getUseCaseByID(ctx *gin.Context, input getUseCaseInputDto) (useCaseEntity, error) {
	useCaseID := uuid.MustParse(input.ID)
	item, err := s.repository.getUseCaseByID(s.storage, useCaseID, false)
	if err != nil {
		return useCaseEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return useCaseEntity{}, errUseCaseNotFound
	}
	return item, nil
}

func (s useCaseService) createUseCase(ctx *gin.Context, input createUseCaseInputDto) (useCaseEntity, error) {
	now := time.Now()
	useCase := useCaseEntity{
		ID:          uuid.New(),
		Title:       input.Title,
		Code:        input.Code,
		Description: input.Description,
		Active:      false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		item, err := s.repository.getUseCaseByCode(tx, input.Code, false)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(item) {
			return errUseCaseSameCodeAlreadyExists
		}
		_, err = s.repository.saveUseCase(tx, useCase)
		if err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseEntity{}, errTransaction
	}
	go s.pubSubAgent.Publish(mm_pubsub.TopicUseCaseV1, mm_pubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mm_pubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mm_pubsub.UseCaseCreatedEvent,
			EventEntity: mm_pubsub.UseCaseEventEntity{
				ID:          useCase.ID,
				Title:       useCase.Title,
				Code:        useCase.Code,
				Description: useCase.Description,
				Active:      useCase.Active,
				CreatedAt:   useCase.CreatedAt,
				UpdatedAt:   useCase.UpdatedAt,
			},
		},
	})
	return useCase, nil
}

func (s useCaseService) updateUseCase(ctx *gin.Context, input updateUseCaseInputDto) (useCaseEntity, error) {
	now := time.Now()
	var useCase useCaseEntity
	err_transaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseId := uuid.MustParse(input.ID)
		item, err := s.repository.getUseCaseByID(tx, useCaseId, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errUseCaseNotFound
		}
		// If the input contains a new code for the use case, check for collision
		if input.Code != nil {
			useCaseSameCode, err := s.repository.getUseCaseByCode(tx, *input.Code, false)
			if err != nil {
				return mm_err.ErrGeneric
			}
			if !mm_utils.IsEmpty(useCaseSameCode) && useCaseSameCode.ID.String() != input.ID {
				return errUseCaseSameCodeAlreadyExists
			}
			// Avoid changing code if the use case is active
			if *input.Code != item.Code && item.Active {
				return errUseCaseCodeChangeNotAllowedWhileActive
			}
		}
		// Update useCase information based on inputs
		useCase = item
		useCase.UpdatedAt = now
		if input.Title != nil {
			useCase.Title = *input.Title
		}
		if input.Description != nil {
			useCase.Description = *input.Description
		}
		if input.Code != nil {
			useCase.Code = *input.Code
		}
		if input.Active != nil {
			// Avoid activating Use Case if there isn't any Fallback Flow
			if !useCase.Active && *input.Active {
				fallbackExists, err := s.repository.checkFallbackFlowExists(tx, useCase.ID)
				if err != nil {
					return mm_err.ErrGeneric
				}
				if !fallbackExists {
					return errUseCaseCannotBeActivatedWithoutFallbackFlow
				}
			}
			useCase.Active = *input.Active
		}
		_, err = s.repository.saveUseCase(tx, useCase)
		if err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if err_transaction != nil {
		return useCaseEntity{}, err_transaction
	}
	// Send an event of useCase updated
	go s.pubSubAgent.Publish(mm_pubsub.TopicUseCaseV1, mm_pubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mm_pubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mm_pubsub.UseCaseUpdatedEvent,
			EventEntity: mm_pubsub.UseCaseEventEntity{
				ID:          useCase.ID,
				Title:       useCase.Title,
				Code:        useCase.Code,
				Description: useCase.Description,
				Active:      useCase.Active,
				CreatedAt:   useCase.CreatedAt,
				UpdatedAt:   useCase.UpdatedAt,
			},
		},
	})
	return useCase, nil
}

func (s useCaseService) deleteUseCase(ctx *gin.Context, input deleteUseCaseInputDto) (useCaseEntity, error) {
	var useCase useCaseEntity
	err_transaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseId := uuid.MustParse(input.ID)
		item, err := s.repository.getUseCaseByID(tx, useCaseId, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errUseCaseNotFound
		}
		// Prevent active use cases from being deleted
		if item.Active {
			return errUseCaseCannotBeDeletedWhileActive
		}
		useCase = item
		s.repository.deleteUseCase(tx, useCase)
		return nil
	})
	if err_transaction != nil {
		return useCaseEntity{}, err_transaction
	}
	// Send an event of useCase updated
	go s.pubSubAgent.Publish(mm_pubsub.TopicUseCaseV1, mm_pubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: mm_pubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: mm_pubsub.UseCaseDeletedEvent,
			EventEntity: mm_pubsub.UseCaseEventEntity{
				ID:          useCase.ID,
				Title:       useCase.Title,
				Code:        useCase.Code,
				Description: useCase.Description,
				Active:      useCase.Active,
				CreatedAt:   useCase.CreatedAt,
				UpdatedAt:   useCase.UpdatedAt,
			},
		},
	})
	return useCase, nil
}
