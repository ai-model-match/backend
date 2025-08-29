package rolloutStrategy

import validation "github.com/go-ozzo/ozzo-validation/v4"

type rsAdaptPhaseDto struct {
	MinFeedback  int64   `json:"min_feedback"`
	MaxStepPct   float64 `json:"max_step_pct"`
	IntervalMins int64   `json:"interval_mins"`
}

func (r rsAdaptPhaseDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.MinFeedback, validation.Required, validation.Min(int64(1))),
		validation.Field(&r.MaxStepPct, validation.Required, validation.Min(float64(1)), validation.Max(float64(100))),
		validation.Field(&r.IntervalMins, validation.Required, validation.Min(int64(1))),
	)
}
