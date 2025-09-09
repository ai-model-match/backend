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
	maxValue := int64(math.MaxInt64)
	newUseCaseStep := useCaseStepEntity{
		ID:          uuid.New(),
		UseCaseID:   useCaseID,
		Title:       input.Title,
		Code:        input.Code,
		Description: input.Description,
		Position:    &maxValue,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		var updatedPosEntities []useCaseStepEntity
		if exists, err := s.repository.checkUseCaseExists(s.storage, useCaseID); err != nil {
			return mm_err.ErrGeneric
		} else if !exists {
			return errUseCaseNotFound
		}
		if item, err := s.repository.getUseCaseStepByCode(tx, useCaseID, input.Code, false); err != nil {
			return mm_err.ErrGeneric
		} else if !mm_utils.IsEmpty(item) {
			return errUseCaseStepSameCodeAlreadyExists
		} else if _, err = s.repository.saveUseCaseStep(tx, newUseCaseStep, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		} else if updatedPosEntities, err = s.repository.recalculateUseCaseStepPosition(tx, useCaseID); err != nil {
			return mm_err.ErrGeneric
		} else if newUseCaseStep, err = s.repository.getUseCaseStepByID(tx, newUseCaseStep.ID, false); err != nil {
			return mm_err.ErrGeneric
		}
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.UseCaseStepCreatedEvent,
				EventEntity: &mm_pubsub.UseCaseStepEventEntity{
					ID:          newUseCaseStep.ID,
					UseCaseID:   newUseCaseStep.UseCaseID,
					Title:       newUseCaseStep.Title,
					Code:        newUseCaseStep.Code,
					Description: newUseCaseStep.Description,
					Position:    newUseCaseStep.Position,
					CreatedAt:   newUseCaseStep.CreatedAt,
					UpdatedAt:   newUseCaseStep.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(useCaseStepEntity{}, newUseCaseStep),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		// For the list of updated entities in Position, send events
		for _, updatedPosEntity := range updatedPosEntities {
			if updatedPosEntity.ID == newUseCaseStep.ID {
				continue
			}
			if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
				Message: mm_pubsub.PubSubEvent{
					EventID:   uuid.New(),
					EventTime: time.Now(),
					EventType: mm_pubsub.UseCaseStepUpdatedEvent,
					EventEntity: &mm_pubsub.UseCaseStepEventEntity{
						ID:          updatedPosEntity.ID,
						UseCaseID:   updatedPosEntity.UseCaseID,
						Title:       updatedPosEntity.Title,
						Code:        updatedPosEntity.Code,
						Description: updatedPosEntity.Description,
						Position:    updatedPosEntity.Position,
						CreatedAt:   updatedPosEntity.CreatedAt,
						UpdatedAt:   updatedPosEntity.UpdatedAt,
					},
					EventChangedFields: []string{"Position", "UpdatedAt"},
				},
			}); err != nil {
				return err
			} else {
				eventsToPublish = append(eventsToPublish, event)
			}
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseStepEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newUseCaseStep, nil
}

func (s useCaseStepService) updateUseCaseStep(ctx *gin.Context, input updateUseCaseStepInputDto) (useCaseStepEntity, error) {
	now := time.Now()
	var updatedUseCaseStep useCaseStepEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case Step exists
		useCaseStepID := uuid.MustParse(input.ID)
		var updatedPosEntities []useCaseStepEntity
		currentUseCaseStep, err := s.repository.getUseCaseStepByID(tx, useCaseStepID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentUseCaseStep) {
			return errUseCaseStepNotFound
		}
		// If the input contains a new code for the use case, check for collision
		if input.Code != nil {
			useCaseStepSameCode, err := s.repository.getUseCaseStepByCode(tx, currentUseCaseStep.UseCaseID, *input.Code, false)
			if err != nil {
				return mm_err.ErrGeneric
			}
			if !mm_utils.IsEmpty(useCaseStepSameCode) && useCaseStepSameCode.ID.String() != input.ID {
				return errUseCaseStepSameCodeAlreadyExists
			}
		}
		// Update useCaseStep information based on inputs
		updatedUseCaseStep = currentUseCaseStep
		updatedUseCaseStep.UpdatedAt = now
		if input.Title != nil {
			updatedUseCaseStep.Title = *input.Title
		}
		if input.Description != nil {
			updatedUseCaseStep.Description = *input.Description
		}
		if input.Code != nil {
			updatedUseCaseStep.Code = *input.Code
		}
		if input.Position != nil {
			// If the step is moving in a lower position (e.g. from 10 to 3),
			// we need to move it one step more, so that, the algorith to re-sort all steps correctly
			if *updatedUseCaseStep.Position > *input.Position {
				*input.Position = *input.Position - 1
			}
			updatedUseCaseStep.Position = input.Position
		}
		if _, err = s.repository.saveUseCaseStep(tx, updatedUseCaseStep, mm_db.Update); err != nil {
			return mm_err.ErrGeneric
		}
		if updatedPosEntities, err = s.repository.recalculateUseCaseStepPosition(tx, updatedUseCaseStep.UseCaseID); err != nil {
			return mm_err.ErrGeneric
		}
		if updatedUseCaseStep, err = s.repository.getUseCaseStepByID(tx, updatedUseCaseStep.ID, false); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of useCaseStep updated
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.UseCaseStepUpdatedEvent,
				EventEntity: &mm_pubsub.UseCaseStepEventEntity{
					ID:          updatedUseCaseStep.ID,
					UseCaseID:   updatedUseCaseStep.UseCaseID,
					Title:       updatedUseCaseStep.Title,
					Code:        updatedUseCaseStep.Code,
					Description: updatedUseCaseStep.Description,
					Position:    updatedUseCaseStep.Position,
					CreatedAt:   updatedUseCaseStep.CreatedAt,
					UpdatedAt:   updatedUseCaseStep.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentUseCaseStep, updatedUseCaseStep),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		// For the list of updated entities in Position, send events
		for _, updatedPosEntity := range updatedPosEntities {
			if updatedPosEntity.ID == updatedUseCaseStep.ID {
				continue
			}
			if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
				Message: mm_pubsub.PubSubEvent{
					EventID:   uuid.New(),
					EventTime: time.Now(),
					EventType: mm_pubsub.UseCaseStepUpdatedEvent,
					EventEntity: &mm_pubsub.UseCaseStepEventEntity{
						ID:          updatedPosEntity.ID,
						UseCaseID:   updatedPosEntity.UseCaseID,
						Title:       updatedPosEntity.Title,
						Code:        updatedPosEntity.Code,
						Description: updatedPosEntity.Description,
						Position:    updatedPosEntity.Position,
						CreatedAt:   updatedPosEntity.CreatedAt,
						UpdatedAt:   updatedPosEntity.UpdatedAt,
					},
					EventChangedFields: []string{"Position", "UpdatedAt"},
				},
			}); err != nil {
				return err
			} else {
				eventsToPublish = append(eventsToPublish, event)
			}
		}
		return nil
	})
	if errTransaction != nil {
		return useCaseStepEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return updatedUseCaseStep, nil
}

func (s useCaseStepService) deleteUseCaseStep(ctx *gin.Context, input deleteUseCaseStepInputDto) (useCaseStepEntity, error) {
	var currentUseCaseStep useCaseStepEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseStepID := uuid.MustParse(input.ID)
		item, err := s.repository.getUseCaseStepByID(tx, useCaseStepID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errUseCaseStepNotFound
		}
		currentUseCaseStep = item
		if _, err := s.repository.deleteUseCaseStep(tx, currentUseCaseStep); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of useCaseStep deleted
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.UseCaseStepDeletedEvent,
				EventEntity: &mm_pubsub.UseCaseStepEventEntity{
					ID:          currentUseCaseStep.ID,
					UseCaseID:   currentUseCaseStep.UseCaseID,
					Title:       currentUseCaseStep.Title,
					Code:        currentUseCaseStep.Code,
					Description: currentUseCaseStep.Description,
					Position:    currentUseCaseStep.Position,
					CreatedAt:   currentUseCaseStep.CreatedAt,
					UpdatedAt:   currentUseCaseStep.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentUseCaseStep, useCaseStepEntity{}),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}

		if updatedPosEntities, err := s.repository.recalculateUseCaseStepPosition(tx, item.UseCaseID); err != nil {
			return mm_err.ErrGeneric
		} else {
			// For the list of updated entities in Position, send events
			for _, updatedPosEntity := range updatedPosEntities {
				if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicUseCaseStepV1, mm_pubsub.PubSubMessage{
					Message: mm_pubsub.PubSubEvent{
						EventID:   uuid.New(),
						EventTime: time.Now(),
						EventType: mm_pubsub.UseCaseStepUpdatedEvent,
						EventEntity: &mm_pubsub.UseCaseStepEventEntity{
							ID:          updatedPosEntity.ID,
							UseCaseID:   updatedPosEntity.UseCaseID,
							Title:       updatedPosEntity.Title,
							Code:        updatedPosEntity.Code,
							Description: updatedPosEntity.Description,
							Position:    updatedPosEntity.Position,
							CreatedAt:   updatedPosEntity.CreatedAt,
							UpdatedAt:   updatedPosEntity.UpdatedAt,
						},
						EventChangedFields: []string{"Position", "UpdatedAt"},
					},
				}); err != nil {
					return err
				} else {
					eventsToPublish = append(eventsToPublish, event)
				}
			}
		}

		return nil
	})
	if errTransaction != nil {
		return useCaseStepEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return currentUseCaseStep, nil
}
