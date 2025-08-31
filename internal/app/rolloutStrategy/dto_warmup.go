package rolloutStrategy

import (
	"errors"

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
	totPct := float64(0)
	for _, goal := range r.Goals {
		totPct = totPct + *mm_utils.RoundTo2Decimals(&goal.FinalServePct)
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

type rsFlowGoalDto struct {
	FlowID        string  `json:"flow_id"`
	FinalServePct float64 `json:"final_serve_pct"`
}

func (r rsFlowGoalDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.FinalServePct, validation.Required, validation.Min(float64(0)), validation.Max(float64(100))),
	)
}
