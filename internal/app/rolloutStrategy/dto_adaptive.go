package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type rsAdaptivePhaseDto struct {
	MinFeedback  int64   `json:"min_feedback"`
	MaxStepPct   float64 `json:"max_step_pct"`
	IntervalMins int64   `json:"interval_mins"`
}

func (r rsAdaptivePhaseDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.MinFeedback, validation.Required, validation.Min(int64(1))),
		validation.Field(&r.MaxStepPct, validation.Required, validation.Min(1.0), validation.Max(100.0)),
		validation.Field(&r.IntervalMins, validation.Required, validation.Min(int64(1))),
	)
}

func (r rsAdaptivePhaseDto) toEntity() mm_pubsub.RsAdaptivePhase {
	return mm_pubsub.RsAdaptivePhase{
		MinFeedback:  r.MinFeedback,
		MaxStepPct:   r.MaxStepPct,
		IntervalMins: r.IntervalMins,
	}
}
