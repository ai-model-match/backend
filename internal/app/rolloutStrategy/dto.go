package rolloutStrategy

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getRolloutStrategyInputDto struct {
	UseCaseID string `uri:"useCaseId"`
}

func (r getRolloutStrategyInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
	)
}

type updateRolloutStrategyInputDto struct {
	UseCaseID     string           `uri:"useCaseId"`
	Configuration rsConfigInputDto `json:"configuration"`
}

func (r updateRolloutStrategyInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.Configuration, validation.Required, validation.By(func(v interface{}) error {
			return v.(rsConfigInputDto).validate()
		})),
	)
}
