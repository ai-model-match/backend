package flowStatistics

import (
	"time"

	"github.com/google/uuid"
)

type flowStatisticsEntity struct {
	ID                 uuid.UUID `json:"id"`
	FlowID             uuid.UUID `json:"flowId"`
	UseCaseID          uuid.UUID `json:"useCaseId"`
	TotRequests        *int64    `json:"totRequests"`
	TotSessionRequests *int64    `json:"totSessionRequests"`
	CurrentServePct    *float64  `json:"currentServePct"`
	TotFeedback        *int64    `json:"totFeedback"`
	AvgScore           *float64  `json:"avgScore"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type flowEntity struct {
	ID        uuid.UUID `json:"flowId"`
	UseCaseID uuid.UUID `json:"useCaseId"`
}
