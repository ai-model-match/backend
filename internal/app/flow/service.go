package flow

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

type flowServiceInterface interface {
	listFlows(ctx *gin.Context, input ListFlowsInputDto) ([]flowEntity, int64, error)
	getFlowByID(ctx *gin.Context, input getFlowInputDto) (flowEntity, error)
	createFlow(ctx *gin.Context, input createFlowInputDto) (flowEntity, error)
	updateFlow(ctx *gin.Context, input updateFlowInputDto) (flowEntity, error)
	deleteFlow(ctx *gin.Context, input deleteFlowInputDto) (flowEntity, error)
	cloneFlow(ctx *gin.Context, input cloneFlowInputDto) (flowEntity, error)
	updateFlowPctBulk(ctx *gin.Context, input updateFlowPctBulkDto) ([]flowEntity, error)
	updateFlowsFromEvent(event mm_pubsub.RsEngineEventEntity) error
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
	newFlow := flowEntity{
		ID:              uuid.New(),
		UseCaseID:       useCaseID,
		Active:          mm_utils.BoolPtr(false),
		Title:           input.Title,
		Description:     input.Description,
		CurrentServePct: mm_utils.Float64Ptr(0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		exists, err := s.repository.checkUseCaseExists(s.storage, useCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !exists {
			return errUseCaseNotFound
		}
		if _, err = s.repository.saveFlow(tx, newFlow, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow created
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowCreatedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:              newFlow.ID,
					UseCaseID:       newFlow.UseCaseID,
					Active:          newFlow.Active,
					Title:           newFlow.Title,
					Description:     newFlow.Description,
					CurrentServePct: newFlow.CurrentServePct,
					CreatedAt:       newFlow.CreatedAt,
					UpdatedAt:       newFlow.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(flowEntity{}, newFlow),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newFlow, nil
}

func (s flowService) updateFlow(ctx *gin.Context, input updateFlowInputDto) (flowEntity, error) {
	now := time.Now()
	var updatedFlow flowEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the Flow exists
		flowID := uuid.MustParse(input.ID)
		currentFlow, err := s.repository.getFlowByID(tx, flowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentFlow) {
			return errFlowNotFound
		}
		// Update flow information based on inputs
		updatedFlow = currentFlow
		updatedFlow.UpdatedAt = now
		if input.Title != nil {
			updatedFlow.Title = *input.Title
		}
		if input.Description != nil {
			updatedFlow.Description = *input.Description
		}
		if input.CurrentServePct != nil {
			updatedFlow.CurrentServePct = mm_utils.RoundTo2DecimalsPtr(input.CurrentServePct)
		}
		if input.Active != nil {
			// If you are trying to deactivate the flow, we need to guarantee that there is at least one active Flow associated to
			// the active Use Case, otherwise return an error
			if *currentFlow.Active && !*input.Active {
				if isUseCaseActive, err := s.repository.checkUseCaseIsActive(tx, updatedFlow.UseCaseID); err != nil {
					return err
				} else if isUseCaseActive {
					if lastActiveFlow, err := s.repository.checkFlowIsLastActive(tx, currentFlow.UseCaseID, currentFlow.ID); err != nil {
						return err
					} else if lastActiveFlow {
						return errFlowCannotBeDeactivatedIfLastActive
					}
				}
			}
			// If you are deactivating the Flow, set its PCT to 0
			if !*input.Active {
				updatedFlow.CurrentServePct = mm_utils.Float64Ptr(0)
			}
			updatedFlow.Active = input.Active
		}

		// Retrieve all the Active Flows
		existingActiveFlows, err := s.repository.getAllActiveFlow(tx, updatedFlow.UseCaseID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		// If this Flow is active and there is not any other active Flow, or this one is the only active, force its PCT to 100%.
		if *updatedFlow.Active {
			if len(existingActiveFlows) == 0 || (len(existingActiveFlows) == 1 && existingActiveFlows[0].ID == updatedFlow.ID) {
				updatedFlow.CurrentServePct = mm_utils.Float64Ptr(100)
			}
		}
		// Save the updated Flow
		if _, err = s.repository.saveFlow(tx, updatedFlow, mm_db.Update); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow updated
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowUpdatedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:              updatedFlow.ID,
					UseCaseID:       updatedFlow.UseCaseID,
					Active:          updatedFlow.Active,
					Title:           updatedFlow.Title,
					Description:     updatedFlow.Description,
					CurrentServePct: updatedFlow.CurrentServePct,
					CreatedAt:       updatedFlow.CreatedAt,
					UpdatedAt:       updatedFlow.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentFlow, updatedFlow),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}

		// Calculate the missing target for all other active Flows (excluding this one)
		target := 100.0
		if *updatedFlow.Active {
			target -= *updatedFlow.CurrentServePct
		}

		// Calculate how much traffic current active flows (excluding this one) are serving
		covered := 0.0
		for _, existingFlow := range existingActiveFlows {
			if existingFlow.ID == updatedFlow.ID {
				continue
			}
			if *existingFlow.CurrentServePct == 0.0 {
				covered += math.SmallestNonzeroFloat64
			} else {
				covered += *existingFlow.CurrentServePct
			}
		}
		// Now calculate the new PCT to serve per each other active Flow
		for _, existingFlow := range existingActiveFlows {
			if existingFlow.ID == updatedFlow.ID {
				continue
			}
			updatedExistingFlow := existingFlow
			if *updatedExistingFlow.CurrentServePct == 0 {
				updatedExistingFlow.CurrentServePct = mm_utils.Float64Ptr(math.SmallestNonzeroFloat64)
			}
			newPct := mm_utils.RoundTo2Decimals(*updatedExistingFlow.CurrentServePct * target / covered)
			updatedExistingFlow.CurrentServePct = &newPct
			updatedExistingFlow.UpdatedAt = time.Now()
			// Save the new PCT of the Flow
			if _, err = s.repository.saveFlow(tx, updatedExistingFlow, mm_db.Update); err != nil {
				return mm_err.ErrGeneric
			}
			// Send an event of flow updated
			if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
				Message: mm_pubsub.PubSubEvent{
					EventID:   uuid.New(),
					EventTime: time.Now(),
					EventType: mm_pubsub.FlowUpdatedEvent,
					EventEntity: &mm_pubsub.FlowEventEntity{
						ID:              updatedExistingFlow.ID,
						UseCaseID:       updatedExistingFlow.UseCaseID,
						Active:          updatedExistingFlow.Active,
						Title:           updatedExistingFlow.Title,
						Description:     updatedExistingFlow.Description,
						CurrentServePct: updatedExistingFlow.CurrentServePct,
						CreatedAt:       updatedExistingFlow.CreatedAt,
						UpdatedAt:       updatedExistingFlow.UpdatedAt,
					},
					EventChangedFields: mm_utils.DiffStructs(existingFlow, updatedExistingFlow),
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
		return flowEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return updatedFlow, nil
}

func (s flowService) deleteFlow(ctx *gin.Context, input deleteFlowInputDto) (flowEntity, error) {
	var currentFlow flowEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		flowID := uuid.MustParse(input.ID)
		currentFlow, err := s.repository.getFlowByID(tx, flowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentFlow) {
			return errFlowNotFound
		}
		// Avoid to delete a Flow if it is active
		if *currentFlow.Active {
			return errFlowCannotBeDeletedIfActive
		}
		if _, err := s.repository.deleteFlow(tx, currentFlow); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow deleted
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowDeletedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:              currentFlow.ID,
					UseCaseID:       currentFlow.UseCaseID,
					Active:          currentFlow.Active,
					Title:           currentFlow.Title,
					Description:     currentFlow.Description,
					CurrentServePct: currentFlow.CurrentServePct,
					CreatedAt:       currentFlow.CreatedAt,
					UpdatedAt:       currentFlow.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentFlow, flowEntity{}),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return currentFlow, nil
}

func (s flowService) cloneFlow(ctx *gin.Context, input cloneFlowInputDto) (flowEntity, error) {
	now := time.Now()
	var newFlow flowEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
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
		newFlow = flowEntity{
			ID:              uuid.New(),
			UseCaseID:       item.UseCaseID,
			Active:          mm_utils.BoolPtr(false),
			Title:           input.NewTitle,
			Description:     item.Description,
			CurrentServePct: mm_utils.Float64Ptr(0),
			CreatedAt:       now,
			UpdatedAt:       now,
			ClonedFromID:    &item.ID,
		}
		if _, err = s.repository.saveFlow(tx, newFlow, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of flow created
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowCreatedEvent,
				EventEntity: &mm_pubsub.FlowEventEntity{
					ID:              newFlow.ID,
					UseCaseID:       newFlow.UseCaseID,
					Active:          newFlow.Active,
					Title:           newFlow.Title,
					Description:     newFlow.Description,
					CurrentServePct: newFlow.CurrentServePct,
					CreatedAt:       newFlow.CreatedAt,
					UpdatedAt:       newFlow.UpdatedAt,
					ClonedFromID:    newFlow.ClonedFromID,
				},
				EventChangedFields: mm_utils.DiffStructs(flowEntity{}, newFlow),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return flowEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newFlow, nil
}

func (s flowService) updateFlowPctBulk(ctx *gin.Context, input updateFlowPctBulkDto) ([]flowEntity, error) {
	useCaseID := uuid.MustParse(input.UseCaseID)
	eventsToPublish := []mm_pubsub.EventToPublish{}
	updatedFlows := []flowEntity{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		exists, err := s.repository.checkUseCaseExists(s.storage, useCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !exists {
			return errUseCaseNotFound
		}
		// Read all active Flows
		var activeFlows []flowEntity
		if activeFlows, err = s.repository.getAllActiveFlow(s.storage, useCaseID, true); err != nil {
			return mm_err.ErrGeneric
		}
		// Prepare indexed Active Flows
		indexedActiveFlows := map[string]flowEntity{}
		for _, activeFlow := range activeFlows {
			indexedActiveFlows[activeFlow.ID.String()] = activeFlow
		}
		// Prepare indexed Input Flows
		indexedInputFlowPcts := map[string]float64{}
		for _, inputFlow := range input.Flows {
			indexedInputFlowPcts[inputFlow.FlowID] = *inputFlow.CurrentServePct
		}
		// Check all inputs are present as active Flows
		for flowID := range indexedInputFlowPcts {
			if _, ok := indexedActiveFlows[flowID]; !ok {
				return errActiveFlowNotFound
			}
		}
		// Loop all active Flows and update their PCTs (default to 0 for missing inputs)
		for flowID, currentFlow := range indexedActiveFlows {
			updatedFlow := currentFlow
			if inputFlowPct, ok := indexedInputFlowPcts[flowID]; ok {
				updatedFlow.CurrentServePct = &inputFlowPct
			} else {
				updatedFlow.CurrentServePct = mm_utils.Float64Ptr(0)
			}
			if _, err = s.repository.saveFlow(tx, updatedFlow, mm_db.Update); err != nil {
				return mm_err.ErrGeneric
			}
			updatedFlow.UpdatedAt = time.Now()
			updatedFlows = append(updatedFlows, updatedFlow)
			// Send an event of flow updated
			if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
				Message: mm_pubsub.PubSubEvent{
					EventID:   uuid.New(),
					EventTime: time.Now(),
					EventType: mm_pubsub.FlowUpdatedEvent,
					EventEntity: &mm_pubsub.FlowEventEntity{
						ID:              updatedFlow.ID,
						UseCaseID:       updatedFlow.UseCaseID,
						Active:          updatedFlow.Active,
						Title:           updatedFlow.Title,
						Description:     updatedFlow.Description,
						CurrentServePct: updatedFlow.CurrentServePct,
						CreatedAt:       updatedFlow.CreatedAt,
						UpdatedAt:       updatedFlow.UpdatedAt,
						ClonedFromID:    updatedFlow.ClonedFromID,
					},
					EventChangedFields: mm_utils.DiffStructs(currentFlow, updatedFlow),
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
		return []flowEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return updatedFlows, nil
}

func (s flowService) updateFlowsFromEvent(event mm_pubsub.RsEngineEventEntity) error {
	eventsToPublish := []mm_pubsub.EventToPublish{}
	updatedFlows := []flowEntity{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		exists, err := s.repository.checkUseCaseExists(s.storage, event.UseCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !exists {
			return errUseCaseNotFound
		}
		// Read all active Flows
		var activeFlows []flowEntity
		if activeFlows, err = s.repository.getAllActiveFlow(s.storage, event.UseCaseID, true); err != nil {
			return mm_err.ErrGeneric
		}
		// Prepare indexed Active Flows
		indexedActiveFlows := map[string]flowEntity{}
		for _, activeFlow := range activeFlows {
			indexedActiveFlows[activeFlow.ID.String()] = activeFlow
		}
		// Prepare indexed Input Flows
		indexedInputFlowPcts := map[string]float64{}
		for _, inputFlow := range event.Flows {
			indexedInputFlowPcts[inputFlow.FlowID.String()] = inputFlow.CurrentServePct
		}
		// Check all inputs are present as active Flows
		for flowID := range indexedInputFlowPcts {
			if _, ok := indexedActiveFlows[flowID]; !ok {
				return errActiveFlowNotFound
			}
		}
		// Loop all active Flows and update their PCTs (default to 0 for missing inputs)
		for flowID, currentFlow := range indexedActiveFlows {
			updatedFlow := currentFlow
			if inputFlowPct, ok := indexedInputFlowPcts[flowID]; ok {
				updatedFlow.CurrentServePct = &inputFlowPct
			} else {
				updatedFlow.CurrentServePct = mm_utils.Float64Ptr(0)
			}
			if _, err = s.repository.saveFlow(tx, updatedFlow, mm_db.Update); err != nil {
				return mm_err.ErrGeneric
			}
			updatedFlow.UpdatedAt = time.Now()
			updatedFlows = append(updatedFlows, updatedFlow)
			// Send an event of flow updated
			if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowV1, mm_pubsub.PubSubMessage{
				Message: mm_pubsub.PubSubEvent{
					EventID:   uuid.New(),
					EventTime: time.Now(),
					EventType: mm_pubsub.FlowUpdatedEvent,
					EventEntity: &mm_pubsub.FlowEventEntity{
						ID:              updatedFlow.ID,
						UseCaseID:       updatedFlow.UseCaseID,
						Active:          updatedFlow.Active,
						Title:           updatedFlow.Title,
						Description:     updatedFlow.Description,
						CurrentServePct: updatedFlow.CurrentServePct,
						CreatedAt:       updatedFlow.CreatedAt,
						UpdatedAt:       updatedFlow.UpdatedAt,
						ClonedFromID:    updatedFlow.ClonedFromID,
					},
					EventChangedFields: mm_utils.DiffStructs(currentFlow, updatedFlow),
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
		return errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return nil
}
