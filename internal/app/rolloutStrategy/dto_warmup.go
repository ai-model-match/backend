package rolloutStrategy

import (
	"errors"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type rsWarmupPhaseDto struct {
	IntervalMins     *int64          `json:"interval_mins"`
	IntervalSessReqs *int64          `json:"interval_sess_req"`
	Goals            []rsFlowGoalDto `json:"goals"`
}

func (r rsWarmupPhaseDto) validate() error {
	if r.IntervalMins == nil && r.IntervalSessReqs == nil {
		return errors.New("at least one between 'interval_mins' or 'interval_sess_req' need to be set")
	}
	if r.IntervalMins != nil && r.IntervalSessReqs != nil {
		return errors.New("only one between 'interval_mins' or 'interval_sess_req' can be set")
	}
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.IntervalMins, validation.Min(int64(1))),
		validation.Field(&r.IntervalSessReqs, validation.Min(int64(1))),
		validation.Field(&r.Goals, validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value interface{}) error {
			v := value.(rsFlowGoalDto)
			return v.validate()
		}))),
	); err != nil {
		return err
	}
	// Check there is only one goal per Flow
	seen := make(map[string]bool)
	totPct := 0.0
	for _, goal := range r.Goals {
		totPct = totPct + *mm_utils.RoundTo2DecimalsPtr(goal.FinalServePct)
		if _, exists := seen[goal.FlowID]; exists {
			return errors.New("flow can have only one goal associated")
		}
		seen[goal.FlowID] = true
	}
	if totPct > 100 {
		return errors.New("goals can have a maximum of 100% as sum")
	}
	return nil
}

func (r rsWarmupPhaseDto) toEntity() mm_pubsub.RsWarmupPhase {
	goals := make([]mm_pubsub.RsFlowGoal, len(r.Goals))
	for i, goal := range r.Goals {
		goals[i] = goal.toEntity()
	}
	return mm_pubsub.RsWarmupPhase{
		IntervalMins:     r.IntervalMins,
		IntervalSessReqs: r.IntervalSessReqs,
		Goals:            goals,
	}
}

type rsFlowGoalDto struct {
	FlowID        string   `json:"flow_id"`
	FinalServePct *float64 `json:"final_serve_pct"`
}

func (r rsFlowGoalDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.FinalServePct, validation.When(r.FinalServePct != nil, validation.Min(0.0), validation.Max(100.0))),
	)
}

func (r rsFlowGoalDto) toEntity() mm_pubsub.RsFlowGoal {
	return mm_pubsub.RsFlowGoal{
		FlowID:        mm_utils.GetUUIDFromString(r.FlowID),
		FinalServePct: *r.FinalServePct,
	}
}
