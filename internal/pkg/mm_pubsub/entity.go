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
	CurrentServePct *float64   `json:"currentServePct"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	ClonedFromID    *uuid.UUID `json:"clonedFromId"`
}

type FlowStatisticsEventEntity struct {
	ID                 uuid.UUID `json:"id"`
	FlowID             uuid.UUID `json:"flowId"`
	UseCaseID          uuid.UUID `json:"useCaseId"`
	TotRequests        int64     `json:"totRequests"`
	TotSessionRequests int64     `json:"totSessionRequests"`
	CurrentServePct    float64   `json:"currentServePct"`
	TotFeedback        int64     `json:"totFeedback"`
	AvgScore           float64   `json:"avgScore"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
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
	Configuration RSConfiguration `json:"configuration"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type RSConfiguration struct {
	Warmup   *RsWarmupPhase  `json:"warmup"`
	Escape   *RsEscapePhase  `json:"escape"`
	Adaptive RsAdaptivePhase `json:"adaptive"`
}

type RsWarmupPhase struct {
	IntervalMins     *int64       `json:"interval_mins"`
	IntervalSessReqs *int64       `json:"interval_sess_req"`
	Goals            []RsFlowGoal `json:"goals"`
}

type RsFlowGoal struct {
	FlowID        string  `json:"flow_id"`
	FinalServePct float64 `json:"final_serve_pct"`
}

type RsEscapePhase struct {
	Rules []RsEscapeRule `json:"rules"`
}

type RsEscapeRule struct {
	FlowID      string             `json:"flow_id"`
	MinFeedback int64              `json:"min_feedback"`
	LowerScore  float64            `json:"lower_score"`
	Rollback    []RsEscapeRollback `json:"rollback"`
}

type RsEscapeRollback struct {
	FlowID        string  `json:"flow_id"`
	FinalServePct float64 `json:"final_serve_pct"`
}

type RsAdaptivePhase struct {
	MinFeedback  int64   `json:"min_feedback"`
	MaxStepPct   float64 `json:"max_step_pct"`
	IntervalMins int64   `json:"interval_mins"`
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

type FeedbackEventEntity struct {
	ID            uuid.UUID `json:"id"`
	UseCaseID     uuid.UUID `json:"useCaseId"`
	FlowID        uuid.UUID `json:"flowId"`
	CorrelationID uuid.UUID `json:"correlationId"`
	Score         float64   `json:"score"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"createdAt"`
}
