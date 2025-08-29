package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type rsConfigInputDto struct {
	Warmup *rsWarmupPhaseDto `json:"warmup"`
	Escape *rsEscapePhaseDto `json:"escape"`
	Adapt  *rsAdaptPhaseDto  `json:"adapt"`
}

func (r rsConfigInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Warmup, validation.By(func(value interface{}) error {
			if mm_utils.IsEmpty(value) {
				return nil
			}
			return value.(*rsWarmupPhaseDto).validate()
		})),
		validation.Field(&r.Escape, validation.By(func(value interface{}) error {
			if mm_utils.IsEmpty(value) {
				return nil
			}
			return value.(*rsEscapePhaseDto).validate()
		})),
		validation.Field(&r.Adapt, validation.By(func(value interface{}) error {
			if mm_utils.IsEmpty(value) {
				return nil
			}
			return value.(*rsAdaptPhaseDto).validate()
		})),
	)
}
