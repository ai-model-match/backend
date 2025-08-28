package flowStatistics

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type flowStatisticsEntity mm_pubsub.FlowStatisticsEventEntity

type flowEntity struct {
	ID        uuid.UUID `json:"flowId"`
	UseCaseID uuid.UUID `json:"useCaseId"`
}
