package feedback

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

type feedbackServiceInterface interface {
	createFeedback(ctx *gin.Context, input createFeedbackInputDto) (feedbackEntity, error)
}

type feedbackService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  feedbackRepositoryInterface
}

func newFeedbackService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository feedbackRepositoryInterface) feedbackService {
	return feedbackService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s feedbackService) createFeedback(ctx *gin.Context, input createFeedbackInputDto) (feedbackEntity, error) {
	now := time.Now()
	var newFeedback feedbackEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		correlation, err := s.repository.getPickerCorrelationByID(tx, uuid.MustParse(input.CorrelationID))
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(correlation) {
			return errCorrelationNotFound
		}
		recentFeedback, err := s.repository.getRecentFeedbackByCorrelationID(tx, uuid.MustParse(input.CorrelationID))
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(recentFeedback) && recentFeedback.CreatedAt.After(correlation.CreatedAt) {
			return errFeedbackAlreadyProvided
		}
		newFeedback = feedbackEntity{
			ID:            uuid.New(),
			CorrelationID: correlation.ID,
			UseCaseID:     correlation.UseCaseID,
			FlowID:        correlation.FlowID,
			Score:         *mm_utils.RoundTo2DecimalsPtr(&input.Score),
			Comment:       input.Comment,
			CreatedAt:     now,
		}
		if _, err = s.repository.saveFeedback(tx, newFeedback, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of feedback created
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicFeedbackV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.FeedbackCreatedEvent,
				EventEntity: &mm_pubsub.FeedbackEventEntity{
					ID:            newFeedback.ID,
					CorrelationID: newFeedback.CorrelationID,
					UseCaseID:     newFeedback.UseCaseID,
					FlowID:        newFeedback.FlowID,
					Score:         newFeedback.Score,
					Comment:       newFeedback.Comment,
					CreatedAt:     newFeedback.CreatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(feedbackEntity{}, newFeedback),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return feedbackEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newFeedback, nil
}
