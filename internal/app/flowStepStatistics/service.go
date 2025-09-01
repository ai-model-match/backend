package flowStepStatistics

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

type flowStepStatisticsServiceInterface interface {
	getFlowStepStatisticsByID(ctx *gin.Context, input getFlowStepStatisticsInputDto) (flowStepStatisticsEntity, error)
	createFlowStepStatistics(flowStepID uuid.UUID) (flowStepStatisticsEntity, error)
	updateStatistics(event mm_pubsub.PickerEventEntity) error
}

type flowStepStatisticsService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  flowStepStatisticsRepositoryInterface
}

func newFlowStepStatisticsService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository flowStepStatisticsRepositoryInterface) flowStepStatisticsService {
	return flowStepStatisticsService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s flowStepStatisticsService) getFlowStepStatisticsByID(ctx *gin.Context, input getFlowStepStatisticsInputDto) (flowStepStatisticsEntity, error) {
	flowStepID := uuid.MustParse(input.FlowStepID)
	item, err := s.repository.getFlowStepStatisticsByFlowStepID(s.storage, flowStepID, false)
	if err != nil {
		return flowStepStatisticsEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return flowStepStatisticsEntity{}, errFlowStepStatisticsNotFound
	}
	return item, nil
}

func (s flowStepStatisticsService) createFlowStepStatistics(flowStepID uuid.UUID) (flowStepStatisticsEntity, error) {
	now := time.Now()
	var flowStepStatistics flowStepStatisticsEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Retrieve the Flow Step and check if exists
		flowStep, err := s.repository.getFlowStepByID(tx, flowStepID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(flowStep) {
			return errFlowNotFound
		}
		// Check if the Flow statistics already exists, if yes, return an error
		item, err := s.repository.getFlowStepStatisticsByFlowStepID(tx, flowStepID, false)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(item) {
			return errFlowStepStatisticsAlreadyExists
		}
		// Create the new Flow Statistics with default values and store it
		flowStepStatistics = flowStepStatisticsEntity{
			ID:          uuid.New(),
			FlowStepID:  flowStep.ID,
			FlowID:      flowStep.FlowID,
			TotRequests: mm_utils.Int64Ptr(0),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if _, err := s.repository.saveFlowStepStatistics(tx, flowStepStatistics, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return flowStepStatisticsEntity{}, errTransaction
	}
	return flowStepStatistics, nil
}

func (s flowStepStatisticsService) updateStatistics(event mm_pubsub.PickerEventEntity) error {
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Find the flow step statistics
		item, err := s.repository.getFlowStepStatisticsByFlowStepID(tx, event.FlowStepID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errFlowStepStatisticsNotFound
		}
		// Update statistics
		*item.TotRequests++
		// And save
		if _, err := s.repository.saveFlowStepStatistics(tx, item, mm_db.Update); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return errTransaction
	}
	return nil
}
