package feedback

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type createFeedbackInputDto struct {
	CorrelationID string  `json:"correlationId"`
	Score         float64 `json:"syntheticScore"`
	Comment       *string `json:"comment"`
	ReferenceLink *string `json:"referenceLink"`
}

func (r createFeedbackInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CorrelationID, validation.Required, is.UUID),
		validation.Field(&r.Score, validation.Required, validation.Min(MinFeedbackScore), validation.Max(MaxFeedbackScore)),
		validation.Field(&r.Comment, validation.Length(0, 4096)),
		validation.Field(&r.ReferenceLink, validation.Length(0, 4096)),
	)
}
