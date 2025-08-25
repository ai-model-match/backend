package useCaseStep

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/google/uuid"
)

type useCaseModel struct {
	ID uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
}

func (m useCaseModel) TableName() string {
	return "mm_use_case"
}

type useCaseStepModel struct {
	ID          uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID   uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	Title       string    `gorm:"column:title;type:varchar(255)"`
	Code        string    `gorm:"column:code;type:varchar(255)"`
	Description string    `gorm:"column:description;type:text"`
	Position    int64     `gorm:"column:position;type:bigint"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m useCaseStepModel) TableName() string {
	return "mm_use_case_step"
}

func (m useCaseStepModel) toEntity() useCaseStepEntity {
	return useCaseStepEntity(m)
}

type useCaseStepOrderBy string

const (
	useCaseStepOrderByPosition  useCaseStepOrderBy = "position"
	useCaseStepOrderByCreatedAt useCaseStepOrderBy = "created_at"
	useCaseStepOrderByUpdatedAt useCaseStepOrderBy = "updated_at"
	useCaseStepOrderByRelevance useCaseStepOrderBy = mm_db.RelevanceField
)

var availableUseCaseStepOrderBy = []interface{}{
	useCaseStepOrderByPosition,
	useCaseStepOrderByCreatedAt,
	useCaseStepOrderByUpdatedAt,
	useCaseStepOrderByRelevance,
}
