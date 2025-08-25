package useCase

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/google/uuid"
)

type useCaseModel struct {
	ID          uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	Title       string    `gorm:"column:title;type:varchar(255)"`
	Code        string    `gorm:"column:code;type:varchar(255)"`
	Description string    `gorm:"column:description;type:text"`
	Active      bool      `gorm:"column:active;type:boolean"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m useCaseModel) TableName() string {
	return "mm_use_case"
}

func (m useCaseModel) toEntity() useCaseEntity {
	return useCaseEntity(m)
}

type flowModel struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	Fallback  bool      `gorm:"column:fallback;type:boolean"`
}

func (m flowModel) TableName() string {
	return "mm_flow"
}

type useCaseOrderBy string

const (
	useCaseOrderByTitle     useCaseOrderBy = "title"
	useCaseOrderByCode      useCaseOrderBy = "code"
	useCaseOrderByActive    useCaseOrderBy = "active"
	useCaseOrderByCreatedAt useCaseOrderBy = "created_at"
	useCaseOrderByUpdatedAt useCaseOrderBy = "updated_at"
	useCaseOrderByRelevance useCaseOrderBy = mm_db.RelevanceField
)

var availableUseCaseOrderBy = []interface{}{
	useCaseOrderByTitle,
	useCaseOrderByCode,
	useCaseOrderByActive,
	useCaseOrderByCreatedAt,
	useCaseOrderByUpdatedAt,
	useCaseOrderByRelevance,
}
