package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type useCaseEntity struct {
	ID     uuid.UUID `json:"id"`
	Active bool      `json:"active"`
}

type flowEntity struct {
	ID              uuid.UUID `json:"id"`
	UseCaseID       uuid.UUID `json:"useCaseId"`
	Active          bool      `json:"active"`
	Fallback        bool      `json:"fallback"`
	CurrentServePct float64   `json:"currentServePct"`
}

type flowStatisticsEntity struct {
	ID                 uuid.UUID `json:"id"`
	FlowID             uuid.UUID `json:"flowId"`
	UseCaseID          uuid.UUID `json:"useCaseId"`
	TotRequests        int64     `json:"totRequests"`
	TotSessionRequests int64     `json:"totSessionRequests"`
	CurrentServePct    float64   `json:"currentServePct"`
	TotFeedback        int64     `json:"totFeedback"`
	AvgScore           float64   `json:"avgScore"`
}

type RolloutState string

type rolloutStrategyEntity struct {
	ID            uuid.UUID                 `json:"id"`
	UseCaseID     uuid.UUID                 `json:"useCaseId"`
	RolloutState  RolloutState              `json:"rolloutState"`
	Configuration mm_pubsub.RSConfiguration `json:"configuration"`
}
