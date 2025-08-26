package flowStep

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_err"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type flowStepServiceInterface interface {
	listFlowSteps(ctx *gin.Context, input ListFlowStepsInputDto) ([]flowStepEntity, int64, error)
	getFlowStepByID(ctx *gin.Context, input getFlowStepInputDto) (flowStepEntity, error)
	updateFlowStep(ctx *gin.Context, input updateFlowStepInputDto) (flowStepEntity, error)
	createStepsForAllFlowsOfUseCase(useCaseID uuid.UUID) error
}

type flowStepService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  flowStepRepositoryInterface
}

func newFlowStepService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository flowStepRepositoryInterface) flowStepService {
	return flowStepService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s flowStepService) listFlowSteps(ctx *gin.Context, input ListFlowStepsInputDto) ([]flowStepEntity, int64, error) {
	flowID := uuid.MustParse(input.FlowID)
	exists, err := s.repository.checkFlowExists(s.storage, flowID)
	if err != nil {
		return []flowStepEntity{}, 0, err
	}
	if !exists {
		return []flowStepEntity{}, 0, errFlowNotFound
	}
	limit, offset := mm_utils.PagePageSizeToLimitOffset(input.Page, input.PageSize)
	items, totalCount, err := s.repository.listFlowSteps(s.storage, flowID, limit, offset, false)
	if err != nil || items == nil {
		return []flowStepEntity{}, 0, err
	}
	return items, totalCount, nil
}

func (s flowStepService) getFlowStepByID(ctx *gin.Context, input getFlowStepInputDto) (flowStepEntity, error) {
	flowStepID := uuid.MustParse(input.ID)
	item, err := s.repository.getFlowStepByID(s.storage, flowStepID, false)
	if err != nil {
		return flowStepEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return flowStepEntity{}, errFlowStepNotFound
	}
	return item, nil
}

func (s flowStepService) updateFlowStep(ctx *gin.Context, input updateFlowStepInputDto) (flowStepEntity, error) {
	now := time.Now()
	var flowStep flowStepEntity
	err_transaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case Step exists
		flowStepID := uuid.MustParse(input.ID)
		if item, err := s.repository.getFlowStepByID(tx, flowStepID, true); err != nil {
			return mm_err.ErrGeneric
		} else if mm_utils.IsEmpty(item) {
			return errFlowStepNotFound
		} else {
			flowStep = item
		}
		if configuration, err := json.Marshal(input.Configuration); err != nil {
			return errFlowStepWrongConfigFormat
		} else {
			flowStep.Configuration = configuration
		}
		// Find placeholders to store
		placeholders := []string{}
		re := regexp.MustCompile(`\\u003c\\u003c([A-Za-z0-9_-]+)\\u003e\\u003e`)
		matches := re.FindAllStringSubmatch(string(flowStep.Configuration), -1)
		for _, match := range matches {
			if len(match) > 1 {
				placeholders = append(placeholders, match[1])
			}
		}
		pl, _ := json.Marshal(placeholders)
		flowStep.Placeholders = json.RawMessage(pl)
		flowStep.UpdatedAt = now
		if _, err := s.repository.saveFlowStep(tx, flowStep, mm_db.Update); err != nil {
			return mm_err.ErrGeneric
		} else if flowStep, err = s.repository.getFlowStepByID(tx, flowStep.ID, false); err != nil {
			return mm_err.ErrGeneric
		} else if err = s.pubSubAgent.Publish(tx, mm_pubsub.TopicFlowStepV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStepUpdatedEvent,
				EventEntity: &mm_pubsub.FlowStepEventEntity{
					ID:            flowStep.ID,
					FlowID:        flowStep.FlowID,
					UseCaseID:     flowStep.UseCaseID,
					UseCaseStepID: flowStep.UseCaseStepID,
					CreatedAt:     flowStep.CreatedAt,
					UpdatedAt:     flowStep.UpdatedAt,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	})
	if err_transaction != nil {
		return flowStepEntity{}, err_transaction
	}

	return flowStep, nil
}

func (s flowStepService) createStepsForAllFlowsOfUseCase(useCaseID uuid.UUID) error {
	now := time.Now()
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		missingFlows, err := s.repository.getAllMissingFlowSteps(s.storage, useCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		for _, missingFlow := range missingFlows {
			config, _ := json.Marshal(map[string]interface{}{})
			placeholders, _ := json.Marshal([]string{})
			flowStep := flowStepEntity{
				ID:            uuid.New(),
				FlowID:        missingFlow.FlowID,
				UseCaseID:     missingFlow.UseCaseID,
				UseCaseStepID: missingFlow.UseCaseStepID,
				Configuration: json.RawMessage(config),
				Placeholders:  json.RawMessage(placeholders),
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if _, err = s.repository.saveFlowStep(tx, flowStep, mm_db.Create); err != nil {
				return mm_err.ErrGeneric
			}
			// Send an event of flowStep created
			if err = s.pubSubAgent.Publish(tx, mm_pubsub.TopicFlowStepV1, mm_pubsub.PubSubMessage{
				Message: mm_pubsub.PubSubEvent{
					EventID:   uuid.New(),
					EventTime: time.Now(),
					EventType: mm_pubsub.FlowStepCreatedEvent,
					EventEntity: &mm_pubsub.FlowStepEventEntity{
						ID:            flowStep.ID,
						FlowID:        flowStep.FlowID,
						UseCaseID:     flowStep.UseCaseID,
						UseCaseStepID: flowStep.UseCaseStepID,
						Configuration: flowStep.Configuration,
						Placeholders:  flowStep.Placeholders,
						CreatedAt:     flowStep.CreatedAt,
						UpdatedAt:     flowStep.UpdatedAt,
					},
				},
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if errTransaction != nil {
		return errTransaction
	}
	return nil
}
