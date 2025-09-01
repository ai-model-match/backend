package mm_pubsub

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UseCaseEventEntity struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Active      *bool     `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UseCaseStepEventEntity struct {
	ID          uuid.UUID `json:"id"`
	UseCaseID   uuid.UUID `json:"useCaseId"`
	Title       string    `json:"title"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Position    *int64    `json:"position"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type FlowEventEntity struct {
	ID              uuid.UUID  `json:"id"`
	UseCaseID       uuid.UUID  `json:"useCaseId"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Active          *bool      `json:"active"`
	Fallback        *bool      `json:"fallback"`
	InitialServePct *float64   `json:"initialServePct"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	ClonedFromID    *uuid.UUID `json:"clonedFromId"`
}

type FlowStepEventEntity struct {
	ID            uuid.UUID       `json:"id"`
	FlowID        uuid.UUID       `json:"flowId"`
	UseCaseID     uuid.UUID       `json:"useCaseId"`
	UseCaseStepID uuid.UUID       `json:"useCaseStepId"`
	Configuration json.RawMessage `json:"configuration"`
	Placeholders  json.RawMessage `json:"placeholders"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type RolloutState string

type RolloutStrategyEventEntity struct {
	ID            uuid.UUID       `json:"id"`
	UseCaseID     uuid.UUID       `json:"useCaseId"`
	RolloutState  RolloutState    `json:"rolloutState"`
	Configuration json.RawMessage `json:"configuration"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type PickerEventEntity struct {
	ID                 uuid.UUID       `json:"id"`
	UseCaseID          uuid.UUID       `json:"useCaseId"`
	UseCaseStepID      uuid.UUID       `json:"useCaseStepId"`
	FlowID             uuid.UUID       `json:"flowId"`
	FlowStepID         uuid.UUID       `json:"flowStepId"`
	CorrelationID      uuid.UUID       `json:"correlationId"`
	IsFirstCorrelation *bool           `json:"IsFirstCorrelation"`
	IsFallback         *bool           `json:"isFallback"`
	InputMessage       json.RawMessage `json:"inputMessage"`
	OutputMessage      json.RawMessage `json:"outputMessage"`
	Placeholders       json.RawMessage `json:"placeholders"`
	CreatedAt          time.Time       `json:"createdAt"`
}
