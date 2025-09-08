package feedback

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type feedbackEntity mm_pubsub.FeedbackEventEntity

type pickerCorrelationEntity struct {
	ID        uuid.UUID
	UseCaseID uuid.UUID
	FlowID    uuid.UUID
	CreatedAt time.Time
}
