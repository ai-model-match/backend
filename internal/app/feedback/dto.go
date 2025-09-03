package feedback

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type createFeedbackInputDto struct {
	CorrelationID string  `json:"correlationId"`
	Score         float64 `json:"score"`
	Comment       string  `json:"comment"`
}

func (r createFeedbackInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CorrelationID, validation.Required, is.UUID),
		validation.Field(&r.Score, validation.Required, validation.Min(MinFeedbackScore), validation.Max(MaxFeedbackScore)),
		validation.Field(&r.Comment, validation.Required, validation.Length(0, 4096)),
	)
}
