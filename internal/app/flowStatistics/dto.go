package flowStatistics

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getFlowStatisticsInputDto struct {
	FlowID string `uri:"flowId"`
}

func (r getFlowStatisticsInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
	)
}
