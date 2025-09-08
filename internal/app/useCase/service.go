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
	newUseCase := useCaseEntity{
		ID:          uuid.New(),
		Title:       input.Title,
		Code:        input.Code,
		Description: input.Description,
		Active:      mm_utils.BoolPtr(false),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		useCaseSameCode, err := s.repository.getUseCaseByCode(tx, input.Code, false)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(useCaseSameCode) {
			return errUseCaseSameCodeAlreadyExists
		}
		_, err = s.repository.saveUseCase(tx, newUseCase, mm_db.Create)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.UseCaseCreatedEvent,
				EventEntity: &mm_pubsub.UseCaseEventEntity{
					ID:          newUseCase.ID,
					Title:       newUseCase.Title,
					Code:        newUseCase.Code,
					Description: newUseCase.Description,
					Active:      newUseCase.Active,
					CreatedAt:   newUseCase.CreatedAt,
					UpdatedAt:   newUseCase.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(useCaseEntity{}, newUseCase),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newUseCase, nil
}

func (s useCaseService) updateUseCase(ctx *gin.Context, input updateUseCaseInputDto) (useCaseEntity, error) {
	now := time.Now()
	var updatedUseCase useCaseEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseId := uuid.MustParse(input.ID)
		currentUseCase, err := s.repository.getUseCaseByID(tx, useCaseId, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentUseCase) {
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
			if *input.Code != currentUseCase.Code && *currentUseCase.Active {
				return errUseCaseCodeChangeNotAllowedWhileActive
			}
		}
		// Update useCase information based on inputs
		updatedUseCase = currentUseCase
		updatedUseCase.UpdatedAt = now
		if input.Title != nil {
			updatedUseCase.Title = *input.Title
		}
		if input.Description != nil {
			updatedUseCase.Description = *input.Description
		}
		if input.Code != nil {
			updatedUseCase.Code = *input.Code
		}
		if input.Active != nil {
			// Avoid activating Use Case if there isn't any Active Flow
			if !*updatedUseCase.Active && *input.Active {
				activeFlowExists, err := s.repository.checkActiveFlowExists(tx, updatedUseCase.ID)
				if err != nil {
					return mm_err.ErrGeneric
				}
				if !activeFlowExists {
					return errUseCaseCannotBeActivatedWithoutActiveFlow
				}
			}
			updatedUseCase.Active = input.Active
		}
		_, err = s.repository.saveUseCase(tx, updatedUseCase, mm_db.Update)
		if err != nil {
			return mm_err.ErrGeneric
		}

		// Send an event of useCase updated
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.UseCaseUpdatedEvent,
				EventEntity: &mm_pubsub.UseCaseEventEntity{
					ID:          updatedUseCase.ID,
					Title:       updatedUseCase.Title,
					Code:        updatedUseCase.Code,
					Description: updatedUseCase.Description,
					Active:      updatedUseCase.Active,
					CreatedAt:   updatedUseCase.CreatedAt,
					UpdatedAt:   updatedUseCase.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentUseCase, updatedUseCase),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return updatedUseCase, nil
}

func (s useCaseService) deleteUseCase(ctx *gin.Context, input deleteUseCaseInputDto) (useCaseEntity, error) {
	var currentUseCase useCaseEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseId := uuid.MustParse(input.ID)
		currentUseCase, err := s.repository.getUseCaseByID(tx, useCaseId, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentUseCase) {
			return errUseCaseNotFound
		}
		// Prevent active use cases from being deleted
		if *currentUseCase.Active {
			return errUseCaseCannotBeDeletedWhileActive
		}
		s.repository.deleteUseCase(tx, currentUseCase)
		// Send an event of useCase deleted
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.UseCaseDeletedEvent,
				EventEntity: &mm_pubsub.UseCaseEventEntity{
					ID:          currentUseCase.ID,
					Title:       currentUseCase.Title,
					Code:        currentUseCase.Code,
					Description: currentUseCase.Description,
					Active:      currentUseCase.Active,
					CreatedAt:   currentUseCase.CreatedAt,
					UpdatedAt:   currentUseCase.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentUseCase, useCaseEntity{}),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return currentUseCase, nil
}
