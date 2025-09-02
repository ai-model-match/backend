package flowStatistics

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

type flowStatisticsServiceInterface interface {
	getFlowStatisticsByID(ctx *gin.Context, input getFlowStatisticsInputDto) (flowStatisticsEntity, error)
	createFlowStatistics(flowID uuid.UUID) (flowStatisticsEntity, error)
	updateStatistics(event mm_pubsub.PickerEventEntity) error
	cleanupStatistics(event mm_pubsub.RolloutStrategyEventEntity) error
}

type flowStatisticsService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  flowStatisticsRepositoryInterface
}

func newFlowStatisticsService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository flowStatisticsRepositoryInterface) flowStatisticsService {
	return flowStatisticsService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s flowStatisticsService) getFlowStatisticsByID(ctx *gin.Context, input getFlowStatisticsInputDto) (flowStatisticsEntity, error) {
	flowID := uuid.MustParse(input.FlowID)
	item, err := s.repository.getFlowStatisticsByFlowID(s.storage, flowID, false)
	if err != nil {
		return flowStatisticsEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return flowStatisticsEntity{}, errFlowStatisticsNotFound
	}
	return item, nil
}

func (s flowStatisticsService) createFlowStatistics(flowID uuid.UUID) (flowStatisticsEntity, error) {
	now := time.Now()
	var flowStatistics flowStatisticsEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Retrieve the Flow and check if exists
		flow, err := s.repository.getFlowByID(tx, flowID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(flow) {
			return errFlowNotFound
		}
		// Check if the Flow statistics already exists, if yes, return an error
		item, err := s.repository.getFlowStatisticsByFlowID(tx, flowID, false)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(item) {
			return errFlowStatisticsAlreadyExists
		}
		// Create the new Flow Statistics with default values and store it
		flowStatistics = flowStatisticsEntity{
			ID:                 uuid.New(),
			FlowID:             flow.ID,
			UseCaseID:          flow.UseCaseID,
			TotRequests:        mm_utils.Int64Ptr(0),
			TotSessionRequests: mm_utils.Int64Ptr(0),
			CurrentServePct:    mm_utils.Float64Ptr(0),
			TotFeedback:        mm_utils.Int64Ptr(0),
			AvgScore:           mm_utils.Float64Ptr(0),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if _, err := s.repository.saveFlowStatistics(tx, flowStatistics, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return flowStatisticsEntity{}, errTransaction
	}
	return flowStatistics, nil
}

func (s flowStatisticsService) updateStatistics(event mm_pubsub.PickerEventEntity) error {
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Find the flow statistics
		item, err := s.repository.getFlowStatisticsByFlowID(tx, event.FlowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(item) {
			return errFlowStatisticsNotFound
		}
		// Update statistics
		*item.TotRequests++
		if *event.IsFirstCorrelation {
			*item.TotSessionRequests++
		}
		// And save
		if _, err := s.repository.saveFlowStatistics(tx, item, mm_db.Update); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return errTransaction
	}
	return nil
}

func (s flowStatisticsService) cleanupStatistics(event mm_pubsub.RolloutStrategyEventEntity) error {
	return s.repository.cleanupFlowStatisticsByUseCaseId(s.storage, event.UseCaseID)
}
