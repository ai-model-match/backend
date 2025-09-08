package flow

import (
	"errors"

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
	ID              string   `uri:"flowId"`
	Title           *string  `json:"title"`
	Description     *string  `json:"description"`
	Active          *bool    `json:"active"`
	CurrentServePct *float64 `json:"currentServePct"`
}

func (r updateFlowInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Title, validation.NilOrNotEmpty, validation.Length(1, 255)),
		validation.Field(&r.Description, validation.NilOrNotEmpty),
		validation.Field(&r.Active, validation.In(true, false)),
		validation.Field(&r.CurrentServePct, validation.Min(0.0), validation.Max(100.0)),
	)
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

type updateFlowPctBulkDto struct {
	UseCaseID string             `json:"useCaseId"`
	Flows     []updateFlowPctDto `json:"flows"`
}

func (r updateFlowPctBulkDto) validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.UseCaseID, validation.Required, is.UUID),
		validation.Field(&r.Flows, validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value interface{}) error {
			v := value.(updateFlowPctDto)
			return v.validate()
		})))); err != nil {
		return err
	}
	// Check there is only one Flow per request and total PCT is 100%
	seen := make(map[string]bool)
	totPct := 0.0
	for _, flow := range r.Flows {
		if flow.CurrentServePct != nil {
			totPct = totPct + *mm_utils.RoundTo2DecimalsPtr(flow.CurrentServePct)
		}
		if _, exists := seen[flow.FlowID]; exists {
			return errors.New("serve PCT can have only one Flow associated")
		}
		seen[flow.FlowID] = true
	}
	if totPct != 100 {
		return errors.New("flows need to reach exactly 100%")
	}
	return nil
}

type updateFlowPctDto struct {
	FlowID          string   `json:"flowId"`
	CurrentServePct *float64 `json:"currentServePct"`
}

func (r updateFlowPctDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.CurrentServePct, validation.When(r.CurrentServePct != nil, validation.Min(0.0), validation.Max(100.0))),
	)
}
