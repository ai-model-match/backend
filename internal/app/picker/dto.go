package picker

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type pickerInputDto struct {
	CorrelationID   string `json:"correlationId"`
	UseCaseCode     string `json:"useCaseCode"`
	UseCaseStepCode string `json:"useCaseStepCode"`
}

func (r pickerInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CorrelationID, validation.Required, is.UUID),
		validation.Field(&r.UseCaseCode, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.UseCaseStepCode, validation.Required, validation.Length(1, 255)),
	)
}
