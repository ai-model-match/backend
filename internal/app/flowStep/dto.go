package flowStep

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ListFlowStepsInputDto struct {
	FlowID   string `form:"flowID"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

func (r ListFlowStepsInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.Page, validation.Required, validation.Min(1)),
		validation.Field(&r.PageSize, validation.Required, validation.Min(1), validation.Max(200)),
	)
}

type getFlowStepInputDto struct {
	ID string `uri:"flowStepID"`
}

func (r getFlowStepInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}

type updateFlowStepInputDto struct {
	ID            string           `uri:"flowStepID"`
	Configuration openAIRequestDTO `json:"configuration"`
}

func (r updateFlowStepInputDto) validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	); err != nil {
		return err
	}
	// Validate nested object
	if err := r.Configuration.validate(); err != nil {
		return err
	}
	return nil
}
