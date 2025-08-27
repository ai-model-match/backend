package flow

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

type flowServiceInterface interface {
	listFlows(ctx *gin.Context, input ListFlowsInputDto) ([]flowEntity, int64, error)
	getFlowByID(ctx *gin.Context, input getFlowInputDto) (flowEntity, error)
	createFlow(ctx *gin.Context, input createFlowInputDto) (flowEntity, error)
	updateFlow(ctx *gin.Context, input updateFlowInputDto) (flowEntity, error)
	deleteFlow(ctx *gin.Context, input deleteFlowInputDto) (flowEntity, error)
	cloneFlow(ctx *gin.Context, input cloneFlowInputDto) (flowEntity, error)
}

type flowService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  flowRepositoryInterface
}

func newFlowService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository flowRepositoryInterface) flowService {
	return flowService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s flowService) listFlows(ctx *gin.Context, input ListFlowsInputDto) ([]flowEntity, int64, error) {
	useCaseID := uuid.MustParse(input.UseCaseID)
	if exists, err := s.repository.checkUseCaseExists(s.storage, useCaseID); err != nil {
		return []flowEntity{}, 0, mm_err.ErrGeneric
	} else if !exists {
		return []flowEntity{}, 0, errUseCaseNotFound
	}
	limit, offset := mm_utils.PagePageSizeToLimitOffset(input.Page, input.PageSize)
	items, totalCount, err := s.repository.listFlows(s.storage, useCaseID, limit, offset, flowOrderBy(input.OrderBy), mm_db.OrderDir(input.OrderDir), input.SearchKey, false)
	if err != nil || items == nil {
		return []flowEntity{}, 0, mm_err.ErrGeneric
	}
	return items, totalCount, nil
}

func (s flowService) getFlowByID(ctx *gin.Context, input getFlowInputDto) (flowEntity, error) {
	flowID := uuid.MustParse(input.ID)
	item, err := s.repository.getFlowByID(s.storage, flowID, false)
	if err != nil {
		return flowEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return flowEntity{}, errFlowNotFound
	}
	return item, nil
}

func (s flowService) createFlow(ctx *gin.Context, input createFlowInputDto) (flowEntity, error) {
	now := time.Now()
	useCaseID := uuid.MustParse(input.UseCaseID)
	flow := flowEntity{
		ID:          uuid.New(),
		UseCaseID:   useCaseID,
		Active:      false,
		Title:       input.Title,
		Description: input.Description,
		Fallback:    false,
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
		if _, err = s.repository.saveFlow(tx, flow, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow created
		if err = s.pubSubAgent.Publish(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowCreatedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:          flow.ID,
					UseCaseID:   flow.UseCaseID,
					Active:      flow.Active,
					Title:       flow.Title,
					Description: flow.Description,
					Fallback:    flow.Fallback,
					CreatedAt:   flow.CreatedAt,
					UpdatedAt:   flow.UpdatedAt,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	}

	return flow, nil
}

func (s flowService) updateFlow(ctx *gin.Context, input updateFlowInputDto) (flowEntity, error) {
	now := time.Now()
	var flow flowEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the Flow exists
		flowID := uuid.MustParse(input.ID)
		item, err := s.repository.getFlowByID(tx, flowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errFlowNotFound
		}
		// Update flow information based on inputs
		flow = item
		flow.UpdatedAt = now
		if input.Title != nil {
			flow.Title = *input.Title
		}
		if input.Description != nil {
			flow.Description = *input.Description
		}
		if input.Active != nil {
			flow.Active = *input.Active
		}
		if input.Fallback != nil {
			// If you are trying to remove the fallback, cannot be done if the use case is active
			if flow.Fallback && !*input.Fallback {
				if isActive, err := s.repository.checkUseCaseIsActive(tx, flow.UseCaseID); err != nil {
					return err
				} else if isActive {
					return errFlowCannotRemoveFallbackWithActiveUseCase
				}
			}
			flow.Fallback = *input.Fallback
		}
		if _, err = s.repository.saveFlow(tx, flow, mm_db.Update); err != nil {
			return mm_err.ErrGeneric
		}
		// If this flow is the fallback one, remove fallback from others if any
		if err = s.repository.makeFallbackConsistent(tx, flow); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow updated
		if err = s.pubSubAgent.Publish(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowUpdatedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:          flow.ID,
					UseCaseID:   flow.UseCaseID,
					Active:      flow.Active,
					Title:       flow.Title,
					Description: flow.Description,
					Fallback:    flow.Fallback,
					CreatedAt:   flow.CreatedAt,
					UpdatedAt:   flow.UpdatedAt,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	}

	return flow, nil
}

func (s flowService) deleteFlow(ctx *gin.Context, input deleteFlowInputDto) (flowEntity, error) {
	var flow flowEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		flowID := uuid.MustParse(input.ID)
		item, err := s.repository.getFlowByID(tx, flowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errFlowNotFound
		}
		flow = item
		// Avoid to delete a Flow if it is a fallback and the use case is active
		if flow.Fallback {
			if isActive, err := s.repository.checkUseCaseIsActive(tx, flow.UseCaseID); err != nil {
				return err
			} else if isActive {
				return errFlowCannotDeleteIfFallbackAndUseCaseActive
			}
		}
		if _, err := s.repository.deleteFlow(tx, flow); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow deleted
		if err = s.pubSubAgent.Publish(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowDeletedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:          flow.ID,
					UseCaseID:   flow.UseCaseID,
					Active:      flow.Active,
					Title:       flow.Title,
					Description: flow.Description,
					Fallback:    flow.Fallback,
					CreatedAt:   flow.CreatedAt,
					UpdatedAt:   flow.UpdatedAt,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	}
	return flow, nil
}

func (s flowService) cloneFlow(ctx *gin.Context, input cloneFlowInputDto) (flowEntity, error) {
	now := time.Now()
	var flow flowEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the flow to be cloned exists
		flowID := uuid.MustParse(input.ID)
		item, err := s.repository.getFlowByID(tx, flowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errFlowNotFound
		}
		// Create a new Flow entity starting from the cloned one
		flow = flowEntity{
			ID:           uuid.New(),
			UseCaseID:    item.UseCaseID,
			Active:       false,
			Title:        input.NewTitle,
			Description:  item.Description,
			Fallback:     false,
			CreatedAt:    now,
			UpdatedAt:    now,
			ClonedFromID: &item.ID,
		}
		if _, err = s.repository.saveFlow(tx, flow, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow created
		if err = s.pubSubAgent.Publish(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowCreatedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:           flow.ID,
					UseCaseID:    flow.UseCaseID,
					Active:       flow.Active,
					Title:        flow.Title,
					Description:  flow.Description,
					Fallback:     flow.Fallback,
					CreatedAt:    flow.CreatedAt,
					UpdatedAt:    flow.UpdatedAt,
					ClonedFromID: flow.ClonedFromID,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	}

	return flow, nil
}
