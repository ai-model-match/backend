package rsEngine

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type flowEntity struct {
	ID              uuid.UUID `json:"id"`
	UseCaseID       uuid.UUID `json:"useCaseId"`
	Active          bool      `json:"active"`
	Fallback        bool      `json:"fallback"`
	CurrentServePct *float64  `json:"currentServePct"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type flowStatisticsEntity struct {
	ID                 uuid.UUID `json:"id"`
	FlowID             uuid.UUID `json:"flowId"`
	UseCaseID          uuid.UUID `json:"useCaseId"`
	TotRequests        int64     `json:"totRequests"`
	TotSessionRequests int64     `json:"totSessionRequests"`
	TotFeedback        int64     `json:"totFeedback"`
	AvgScore           float64   `json:"avgScore"`
}

type rolloutStrategyEntity struct {
	ID            uuid.UUID                 `json:"id"`
	UseCaseID     uuid.UUID                 `json:"useCaseId"`
	RolloutState  mm_pubsub.RolloutState    `json:"rolloutState"`
	Configuration mm_pubsub.RSConfiguration `json:"configuration"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
}
