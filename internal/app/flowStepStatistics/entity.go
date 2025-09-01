package flowStepStatistics

import (
	"time"

	"github.com/google/uuid"
)

type flowStepStatisticsEntity struct {
	ID          uuid.UUID `json:"id"`
	FlowStepID  uuid.UUID `json:"flowStepId"`
	FlowID      uuid.UUID `json:"flowId"`
	TotRequests *int64    `json:"totRequests"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type flowStepEntity struct {
	ID     uuid.UUID `json:"flowStepId"`
	FlowID uuid.UUID `json:"flowId"`
}
