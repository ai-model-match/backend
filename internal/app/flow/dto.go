package flow

import (
	"strconv"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ListFlowsInputDto struct {
	UseCaseID string  `form:"useCaseId"`
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
	ID string `uri:"flowId"`
}

func (r getFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}

type createFlowInputDto struct {
	UseCaseID   string `json:"useCaseId"`
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
	ID                    string  `uri:"flowId"`
	Title                 *string `json:"title"`
	Description           *string `json:"description"`
	Active                *bool   `json:"active"`
	Fallback              *bool   `json:"fallback"`
	InitialServePctString *string `json:"initialServePct"`
	InitialServePct       *float64
}

func (r updateFlowInputDto) validate() (updateFlowInputDto, error) {
	// Transform string input as float 64
	if r.InitialServePctString != nil {
		if initialServePct, err := strconv.ParseFloat(*r.InitialServePctString, 64); err != nil {
			return updateFlowInputDto{}, validation.ErrMatchInvalid
		} else {
			r.InitialServePct = &initialServePct
		}
	}
	// Validate and return the value
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Title, validation.NilOrNotEmpty, validation.Length(1, 255)),
		validation.Field(&r.Description, validation.NilOrNotEmpty),
		validation.Field(&r.Active, validation.In(true, false)),
		validation.Field(&r.Fallback, validation.In(true, false)),
		validation.Field(&r.InitialServePct, validation.When(r.InitialServePct != nil, validation.Min(float64(0)), validation.Max(float64(100)))),
	); err != nil {
		return updateFlowInputDto{}, err
	}
	return r, nil
}

type deleteFlowInputDto struct {
	ID string `uri:"flowId"`
}

func (r deleteFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}

type cloneFlowInputDto struct {
	ID       string `uri:"flowId"`
	NewTitle string `json:"newTitle"`
}

func (r cloneFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.NewTitle, validation.NilOrNotEmpty, validation.Length(1, 255)),
	)
}
