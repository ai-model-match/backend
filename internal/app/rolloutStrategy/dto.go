package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
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
	UseCaseID     string            `uri:"useCaseId"`
	RolloutState  *string           `json:"state"`
	Configuration *rsConfigInputDto `json:"configuration"`
}

func (r updateRolloutStrategyInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.RolloutState, validation.In(mm_utils.TransformToStrings(AvailableRolloutState)...)),
		validation.Field(&r.Configuration, validation.When(r.RolloutState != nil, validation.Nil.Error("not allowed during State update")), validation.By(func(v interface{}) error {
			if mm_utils.IsEmpty(v) {
				return nil
			}
			return v.(*rsConfigInputDto).validate()
		})),
	)
}
