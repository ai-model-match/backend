package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
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
	UseCaseID     string           `uri:"useCaseId"`
	Configuration rsConfigInputDto `json:"configuration"`
}

func (r updateRolloutStrategyInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.Configuration, validation.By(func(v interface{}) error {
			return v.(rsConfigInputDto).validate()
		})),
	)
}

type updateRolloutStrategyStatusInputDto struct {
	UseCaseID    string `uri:"useCaseId"`
	RolloutState string `json:"state"`
}

func (r updateRolloutStrategyStatusInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.RolloutState, validation.Required, validation.In(mm_utils.TransformToStrings(mm_pubsub.AvailableRolloutState)...)),
	)
}
