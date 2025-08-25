package flow

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ListFlowsInputDto struct {
	UseCaseID string  `form:"useCaseID"`
	Page      int     `form:"page"`
	PageSize  int     `form:"pageSize"`
	OrderBy   string  `form:"orderBy"`
	OrderDir  string  `form:"orderDir"`
	SearchKey *string `form:"searchKey"`
}

func (r ListFlowsInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.Page, validation.Required, validation.Min(1)),
		validation.Field(&r.PageSize, validation.Required, validation.Min(1), validation.Max(200)),
		validation.Field(&r.OrderBy, validation.Required, validation.In(mm_utils.TransformToStrings(availableFlowOrderBy)...), validation.When(r.SearchKey == nil, validation.NotIn(mm_db.RelevanceField).Error("not allowed without searchKey field"))),
		validation.Field(&r.OrderDir, validation.Required, validation.In(mm_utils.TransformToStrings(mm_db.AvailableOrderDir)...)),
		validation.Field(&r.SearchKey, validation.Length(3, 200)),
	)
}

type getFlowInputDto struct {
	ID string `uri:"flowID"`
}

func (r getFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}

type createFlowInputDto struct {
	UseCaseID   string `json:"useCaseID"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (r createFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.Title, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.Description, validation.Required),
	)
}

type updateFlowInputDto struct {
	ID          string  `uri:"flowID"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
	Fallback    *bool   `json:"fallback"`
}

func (r updateFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Title, validation.NilOrNotEmpty, validation.Length(1, 255)),
		validation.Field(&r.Description, validation.NilOrNotEmpty),
		validation.Field(&r.Active, validation.In(true, false)),
		validation.Field(&r.Fallback, validation.In(true, false)),
	)
}

type deleteFlowInputDto struct {
	ID string `uri:"flowID"`
}

func (r deleteFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}
