package rolloutStrategy

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type rsEscapePhaseDto struct {
	Rules []rsEscapeRulesDto `json:"rules"`
}

func (r rsEscapePhaseDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Rules, validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value interface{}) error {
			v := value.(rsEscapeRulesDto)
			return v.validate()
		}))),
	)
}

type rsEscapeRulesDto struct {
	FlowID      string  `json:"flow_id"`
	MinFeedback int64   `json:"min_feedback"`
	LowerScore  float64 `json:"lower_score"`
}

func (r rsEscapeRulesDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.MinFeedback, validation.Required, validation.Min(int64(1))),
		validation.Field(&r.LowerScore, validation.Required, validation.Min(MinFeedbackScore), validation.Max(MaxFeedbackScore)),
	)
}
