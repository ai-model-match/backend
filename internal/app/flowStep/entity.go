package flowStep

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type flowStepEntity mm_pubsub.FlowStepEventEntity

type missingFlowStepEntity struct {
	FlowID        uuid.UUID `json:"flowID"`
	UseCaseID     uuid.UUID `json:"useCaseId"`
	UseCaseStepID uuid.UUID `json:"useCaseStepId"`
}
