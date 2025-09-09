package picker

import (
	"encoding/json"
	"math/rand/v2"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_err"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pickerServiceInterface interface {
	pick(ctx *gin.Context, input pickerInputDto) (pickerEntity, error)
}

type pickerService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  pickerRepositoryInterface
}

func newPickerService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository pickerRepositoryInterface) pickerService {
	return pickerService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s pickerService) pick(ctx *gin.Context, input pickerInputDto) (pickerEntity, error) {
	var useCase useCaseEntity
	var useCaseStep useCaseStepEntity
	var correlation pickerCorrelationEntity
	var availableFlows []flowEntity
	var selectedFlow flowEntity
	var selectedFlowStep flowStepEntity
	var newPickedEntity pickerEntity
	var isFirstCorrelation bool = false
	eventsToPublish := []mm_pubsub.EventToPublish{}
	// Check Use Case exists by its code
	if item, err := s.repository.getUseCaseByCode(s.storage, input.UseCaseCode); err != nil {
		return pickerEntity{}, mm_err.ErrGeneric
	} else if mm_utils.IsEmpty(item) {
		return pickerEntity{}, errUseCaseNotFound
	} else if !item.Active {
		return pickerEntity{}, errUseCaseNotAcive
	} else {
		useCase = item
	}
	// Check Use Case Step exists by its code and associated to the Use Case before
	if item, err := s.repository.getUseCaseStepByCode(s.storage, useCase.ID, input.UseCaseStepCode); err != nil {
		return pickerEntity{}, mm_err.ErrGeneric
	} else if mm_utils.IsEmpty(item) {
		return pickerEntity{}, errUseCaseStepNotFound
	} else {
		useCaseStep = item
	}
	// Search a recent correlation by ID
	if item, err := s.repository.getRecentCorrelationByID(s.storage, mm_utils.GetUUIDFromString(input.CorrelationID)); err != nil {
		return pickerEntity{}, mm_err.ErrGeneric
	} else if !mm_utils.IsEmpty(item) {
		if item.UseCaseID != useCase.ID {
			return pickerEntity{}, errCorrelationConflict
		}
		correlation = item
	}
	if !mm_utils.IsEmpty(correlation) {
		// If correlation found, we have immediately the Flow
		if item, err := s.repository.getFlowByID(s.storage, correlation.FlowID); err != nil {
			return pickerEntity{}, mm_err.ErrGeneric
		} else if mm_utils.IsEmpty(item) {
			return pickerEntity{}, errFlowNotFound
		} else {
			selectedFlow = item
		}
	} else {
		// If correlation does not exist, retrieve all flows related to the Use Case
		if items, err := s.repository.getFlowsByUseCaseID(s.storage, useCase.ID); err != nil {
			return pickerEntity{}, mm_err.ErrGeneric
		} else if len(items) == 0 {
			return pickerEntity{}, errFlowsNotAvailable
		} else {
			// Prepare list of active Flows to consider
			for _, item := range items {
				if item.Active {
					availableFlows = append(availableFlows, item)
				}
			}
		}
		// All available flows are expected to add up to 100%, but due to rounding they may be slightly off.
		// To handle this safely, we use a weighted random selection.
		totalPct := 0.0
		for i := range availableFlows {
			totalPct += availableFlows[i].CurrentServePct
		}
		r := rand.Float64() * totalPct
		var cumulative float64
		for i := range availableFlows {
			cumulative += availableFlows[i].CurrentServePct
			if r < cumulative {
				selectedFlow = availableFlows[i]
				break
			}
		}
	}
	// Retrieve the Step of the selected Flow
	if item, err := s.repository.getFlowStepByFlowIdandUseCaseStepId(s.storage, selectedFlow.ID, useCaseStep.ID); err != nil {
		return pickerEntity{}, mm_err.ErrGeneric
	} else if mm_utils.IsEmpty(item) {
		return pickerEntity{}, errUseCaseStepNotFound
	} else {
		selectedFlowStep = item
	}

	// Start transaction
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Now store the correlation for next requests, updating old ones if needed
		if mm_utils.IsEmpty(correlation) {
			correlation = pickerCorrelationEntity{
				ID:        mm_utils.GetUUIDFromString(input.CorrelationID),
				UseCaseID: useCase.ID,
				FlowID:    selectedFlow.ID,
				CreatedAt: time.Now(),
			}
			if _, err := s.repository.saveCorrelation(s.storage, correlation, mm_db.Upsert); err != nil {
				return mm_err.ErrGeneric
			} else {
				isFirstCorrelation = true
			}
		}
		// Prepare response and event
		inputMsg, err := json.Marshal(input)
		if err != nil {
			return mm_err.ErrGeneric
		}
		newPickedEntity = pickerEntity{
			ID:                 uuid.New(),
			UseCaseID:          useCase.ID,
			UseCaseStepID:      useCaseStep.ID,
			FlowID:             selectedFlow.ID,
			FlowStepID:         selectedFlowStep.ID,
			CorrelationID:      mm_utils.GetUUIDFromString(input.CorrelationID),
			IsFirstCorrelation: &isFirstCorrelation,
			InputMessage:       inputMsg,
			OutputMessage:      selectedFlowStep.Configuration,
			Placeholders:       selectedFlowStep.Placeholders,
			CreatedAt:          time.Now(),
		}
		// Save request and relative response on DB for further analysis
		if _, err := s.repository.savePickerEntity(tx, newPickedEntity, mm_db.Create); err != nil {
			return err
		}
		// Persist an event to Picker topic
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicPickerV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.PickerMatchedEvent,
				EventEntity: &mm_pubsub.PickerEventEntity{
					ID:                 newPickedEntity.ID,
					UseCaseID:          newPickedEntity.UseCaseID,
					UseCaseStepID:      newPickedEntity.UseCaseStepID,
					FlowID:             newPickedEntity.FlowID,
					FlowStepID:         newPickedEntity.FlowStepID,
					CorrelationID:      newPickedEntity.CorrelationID,
					IsFirstCorrelation: newPickedEntity.IsFirstCorrelation,
					InputMessage:       newPickedEntity.InputMessage,
					OutputMessage:      newPickedEntity.OutputMessage,
					Placeholders:       newPickedEntity.Placeholders,
					CreatedAt:          newPickedEntity.CreatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(pickerEntity{}, newPickedEntity),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return pickerEntity{}, errTransaction
	} else {
		// Send event on PubSub
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newPickedEntity, nil

}
