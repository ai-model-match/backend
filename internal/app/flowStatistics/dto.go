package flowStatistics

import (
	"strconv"

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

type updateFlowStatisticsInputDto struct {
	FlowID                string  `uri:"flowId"`
	FlowStatisticsID      string  `uri:"flowStatisticsId"`
	InitialServePctString *string `json:"initialServePct"`
	InitialServePct       *float64
}

func (r updateFlowStatisticsInputDto) validate() (updateFlowStatisticsInputDto, error) {
	// Transform string input as float 64
	if r.InitialServePctString != nil {
		if initialServePct, err := strconv.ParseFloat(*r.InitialServePctString, 64); err != nil {
			return updateFlowStatisticsInputDto{}, validation.ErrMatchInvalid
		} else {
			r.InitialServePct = &initialServePct
		}
	}
	// Validate and return the value
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.FlowStatisticsID, validation.Required, is.UUID),
		validation.Field(&r.InitialServePct, validation.When(r.InitialServePct != nil, validation.Min(float64(0)), validation.Max(float64(100)))),
	); err != nil {
		return updateFlowStatisticsInputDto{}, err
	}
	return r, nil
}
