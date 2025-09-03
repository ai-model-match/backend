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
	var flowStatistics flowStatisticsEntity
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
		flowStatistics = flowStatisticsEntity{
			ID:                 uuid.New(),
			FlowID:             flow.ID,
			UseCaseID:          flow.UseCaseID,
			TotRequests:        0,
			TotSessionRequests: 0,
			CurrentServePct:    0,
			TotFeedback:        0,
			AvgScore:           0,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if _, err := s.repository.saveFlowStatistics(tx, flowStatistics, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Persist event
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowStatisticsV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStatisticsCreatedEvent,
				EventEntity: &mm_pubsub.FlowStatisticsEventEntity{
					ID:                 flowStatistics.ID,
					FlowID:             flowStatistics.FlowID,
					UseCaseID:          flowStatistics.UseCaseID,
					TotRequests:        flowStatistics.TotRequests,
					TotSessionRequests: flowStatistics.TotSessionRequests,
					CurrentServePct:    flowStatistics.CurrentServePct,
					TotFeedback:        flowStatistics.TotFeedback,
					AvgScore:           flowStatistics.AvgScore,
					CreatedAt:          flowStatistics.CreatedAt,
					UpdatedAt:          flowStatistics.UpdatedAt,
				},
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
	return flowStatistics, nil
}

func (s flowStatisticsService) updateRequestStatistics(event mm_pubsub.PickerEventEntity) error {
	eventsToPublish := []mm_pubsub.EventToPublish{}
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
		item.TotRequests++
		if *event.IsFirstCorrelation {
			item.TotSessionRequests++
		}
		// And save
		if _, err := s.repository.saveFlowStatistics(tx, item, mm_db.Update); err != nil {
			return err
		}
		// Persist event
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowStatisticsV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStatisticsUpdatedEvent,
				EventEntity: &mm_pubsub.FlowStatisticsEventEntity{
					ID:                 item.ID,
					FlowID:             item.FlowID,
					UseCaseID:          item.UseCaseID,
					TotRequests:        item.TotRequests,
					TotSessionRequests: item.TotSessionRequests,
					CurrentServePct:    item.CurrentServePct,
					TotFeedback:        item.TotFeedback,
					AvgScore:           item.AvgScore,
					CreatedAt:          item.CreatedAt,
					UpdatedAt:          item.UpdatedAt,
				},
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
		item.TotFeedback++
		newAvg := ((item.AvgScore * float64(item.TotFeedback-1)) + event.Score) / float64(item.TotFeedback)
		item.AvgScore = *mm_utils.RoundTo2DecimalsPtr(&newAvg)
		// And save
		if _, err := s.repository.saveFlowStatistics(tx, item, mm_db.Update); err != nil {
			return err
		}
		// Persist event
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFlowStatisticsV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FlowStatisticsUpdatedEvent,
				EventEntity: &mm_pubsub.FlowStatisticsEventEntity{
					ID:                 item.ID,
					FlowID:             item.FlowID,
					UseCaseID:          item.UseCaseID,
					TotRequests:        item.TotRequests,
					TotSessionRequests: item.TotSessionRequests,
					CurrentServePct:    item.CurrentServePct,
					TotFeedback:        item.TotFeedback,
					AvgScore:           item.AvgScore,
					CreatedAt:          item.CreatedAt,
					UpdatedAt:          item.UpdatedAt,
				},
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
	return s.repository.cleanupFlowStatisticsByUseCaseId(s.storage, event.UseCaseID)
}
