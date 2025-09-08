package picker

import (
	"encoding/json"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type useCaseEntity struct {
	ID     uuid.UUID
	Code   string
	Active bool
}

type useCaseStepEntity struct {
	ID   uuid.UUID
	Code string
}

type flowEntity struct {
	ID              uuid.UUID
	UseCaseID       uuid.UUID
	Active          bool
	CurrentServePct float64
}

type flowStepEntity struct {
	ID            uuid.UUID
	FlowID        uuid.UUID
	UseCaseID     uuid.UUID
	UseCaseStepID uuid.UUID
	Configuration json.RawMessage
	Placeholders  json.RawMessage
}

type pickerCorrelationEntity struct {
	ID        uuid.UUID
	UseCaseID uuid.UUID
	FlowID    uuid.UUID
	CreatedAt time.Time
}

type pickerEntity mm_pubsub.PickerEventEntity
