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
	updateRequestStatistics(event mm_pubsub.PickerEventEntity) error
	updateFeedbackStatistics(event mm_pubsub.FeedbackEventEntity) error
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
	var newFlowStatistics flowStatisticsEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
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
		newFlowStatistics = flowStatisticsEntity{
			ID:                 uuid.New(),
			FlowID:             flow.ID,
			UseCaseID:          flow.UseCaseID,
			TotRequests:        0,
			TotSessionRequests: 0,
			TotFeedback:        0,
			AvgScore:           0,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if _, err := s.repository.saveFlowStatistics(tx, newFlowStatistics, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Persist event
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowStatisticsV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStatisticsCreatedEvent,
				EventEntity: &mm_pubsub.FlowStatisticsEventEntity{
					ID:                 newFlowStatistics.ID,
					FlowID:             newFlowStatistics.FlowID,
					UseCaseID:          newFlowStatistics.UseCaseID,
					TotRequests:        newFlowStatistics.TotRequests,
					TotSessionRequests: newFlowStatistics.TotSessionRequests,
					TotFeedback:        newFlowStatistics.TotFeedback,
					AvgScore:           newFlowStatistics.AvgScore,
					CreatedAt:          newFlowStatistics.CreatedAt,
					UpdatedAt:          newFlowStatistics.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(flowStatisticsEntity{}, newFlowStatistics),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return flowStatisticsEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newFlowStatistics, nil
}

func (s flowStatisticsService) updateRequestStatistics(event mm_pubsub.PickerEventEntity) error {
	eventsToPublish := []mm_pubsub.EventToPublish{}
	var updatedFlowStatistics flowStatisticsEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Find the flow statistics
		currentFlowStatistics, err := s.repository.getFlowStatisticsByFlowID(tx, event.FlowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentFlowStatistics) {
			return errFlowStatisticsNotFound
		}
		// Update statistics
		updatedFlowStatistics = currentFlowStatistics
		updatedFlowStatistics.TotRequests++
		if *event.IsFirstCorrelation {
			updatedFlowStatistics.TotSessionRequests++
		}
		// And save
		if _, err := s.repository.saveFlowStatistics(tx, updatedFlowStatistics, mm_db.Update); err != nil {
			return err
		}
		// Persist event
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowStatisticsV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStatisticsUpdatedEvent,
				EventEntity: &mm_pubsub.FlowStatisticsEventEntity{
					ID:                 updatedFlowStatistics.ID,
					FlowID:             updatedFlowStatistics.FlowID,
					UseCaseID:          updatedFlowStatistics.UseCaseID,
					TotRequests:        updatedFlowStatistics.TotRequests,
					TotSessionRequests: updatedFlowStatistics.TotSessionRequests,
					TotFeedback:        updatedFlowStatistics.TotFeedback,
					AvgScore:           updatedFlowStatistics.AvgScore,
					CreatedAt:          updatedFlowStatistics.CreatedAt,
					UpdatedAt:          updatedFlowStatistics.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentFlowStatistics, updatedFlowStatistics),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
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

func (s flowStatisticsService) updateFeedbackStatistics(event mm_pubsub.FeedbackEventEntity) error {
	eventsToPublish := []mm_pubsub.EventToPublish{}
	var updatedFlowStatistics flowStatisticsEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Find the flow statistics
		currentFlowStatistics, err := s.repository.getFlowStatisticsByFlowID(tx, event.FlowID, true)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(currentFlowStatistics) {
			return errFlowStatisticsNotFound
		}
		// Update statistics
		updatedFlowStatistics = currentFlowStatistics
		updatedFlowStatistics.TotFeedback++
		newAvg := ((updatedFlowStatistics.AvgScore * float64(updatedFlowStatistics.TotFeedback-1)) + event.Score) / float64(updatedFlowStatistics.TotFeedback)
		updatedFlowStatistics.AvgScore = *mm_utils.RoundTo2DecimalsPtr(&newAvg)
		// And save
		if _, err := s.repository.saveFlowStatistics(tx, updatedFlowStatistics, mm_db.Update); err != nil {
			return err
		}
		// Persist event
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowStatisticsV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStatisticsUpdatedEvent,
				EventEntity: &mm_pubsub.FlowStatisticsEventEntity{
					ID:                 updatedFlowStatistics.ID,
					FlowID:             updatedFlowStatistics.FlowID,
					UseCaseID:          updatedFlowStatistics.UseCaseID,
					TotRequests:        updatedFlowStatistics.TotRequests,
					TotSessionRequests: updatedFlowStatistics.TotSessionRequests,
					TotFeedback:        updatedFlowStatistics.TotFeedback,
					AvgScore:           updatedFlowStatistics.AvgScore,
					CreatedAt:          updatedFlowStatistics.CreatedAt,
					UpdatedAt:          updatedFlowStatistics.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentFlowStatistics, updatedFlowStatistics),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
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

func (s flowStatisticsService) cleanupStatistics(event mm_pubsub.RolloutStrategyEventEntity) error {
	// If needed, send a new cleanup event for each Flow Statistics impacted
	return s.repository.cleanupFlowStatisticsByUseCaseId(s.storage, event.UseCaseID)
}
