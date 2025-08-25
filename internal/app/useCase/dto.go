package useCase

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ListUseCasesInputDto struct {
	Page      int     `form:"page"`
	PageSize  int     `form:"pageSize"`
	OrderBy   string  `form:"orderBy"`
	OrderDir  string  `form:"orderDir"`
	SearchKey *string `form:"searchKey"`
}

func (r ListUseCasesInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Page, validation.Required, validation.Min(1)),
		validation.Field(&r.PageSize, validation.Required, validation.Min(1), validation.Max(200)),
		validation.Field(&r.OrderBy, validation.Required, validation.In(mm_utils.TransformToStrings(availableUseCaseOrderBy)...), validation.When(r.SearchKey == nil, validation.NotIn(mm_db.RelevanceField).Error("not allowed without searchKey field"))),
		validation.Field(&r.OrderDir, validation.Required, validation.In(mm_utils.TransformToStrings(mm_db.AvailableOrderDir)...)),
		validation.Field(&r.SearchKey, validation.Length(3, 200)),
	)
}

type getUseCaseInputDto struct {
	ID string `uri:"useCaseID"`
}

func (r getUseCaseInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}

type createUseCaseInputDto struct {
	Title       string `json:"title"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (r createUseCaseInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.Code, validation.Required, validation.Length(1, 255)),
		validation.Field(&r.Description, validation.Required),
	)
}

type updateUseCaseInputDto struct {
	ID          string  `uri:"useCaseID"`
	Title       *string `json:"title"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
}

func (r updateUseCaseInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Title, validation.NilOrNotEmpty, validation.Length(1, 255)),
		validation.Field(&r.Code, validation.NilOrNotEmpty, validation.Length(1, 255)),
		validation.Field(&r.Description, validation.NilOrNotEmpty),
	)
}

type deleteUseCaseInputDto struct {
	ID string `uri:"useCaseID"`
}

func (r deleteUseCaseInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}
