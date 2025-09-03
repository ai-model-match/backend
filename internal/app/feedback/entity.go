package feedback

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type feedbackEntity mm_pubsub.FeedbackEventEntity

type pickerCorrelationEntity struct {
	ID        uuid.UUID `json:"id"`
	UseCaseID uuid.UUID `json:"useCaseId"`
	FlowID    uuid.UUID `json:"flowId"`
	Fallback  bool      `json:"fallback"`
}
