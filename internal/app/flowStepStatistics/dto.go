package flowStepStatistics

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getFlowStepStatisticsInputDto struct {
	FlowStepID string `uri:"flowStepId"`
}

func (r getFlowStepStatisticsInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowStepID, validation.Required, is.UUID),
	)
}
