package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type rsConfigInputDto struct {
	Warmup   *rsWarmupPhaseDto  `json:"warmup"`
	Escape   *rsEscapePhaseDto  `json:"escape"`
	Adaptive rsAdaptivePhaseDto `json:"adaptive"`
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
		validation.Field(&r.Adaptive, validation.By(func(value interface{}) error {
			return value.(rsAdaptivePhaseDto).validate()
		})),
	)
}

func (r rsConfigInputDto) toEntity() mm_pubsub.RSConfiguration {
	a := mm_pubsub.RSConfiguration{
		Adaptive: r.Adaptive.toEntity(),
	}
	if r.Warmup != nil {
		e := r.Warmup.toEntity()
		a.Warmup = &e
	}
	if r.Escape != nil {
		e := r.Escape.toEntity()
		a.Escape = &e
	}
	return a
}
