package flow

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/google/uuid"
)

type useCaseModel struct {
	ID     uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	Active *bool     `gorm:"column:active;type:boolean"`
}

func (m useCaseModel) TableName() string {
	return "mm_use_case"
}

type flowModel struct {
	ID              uuid.UUID  `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID       uuid.UUID  `gorm:"column:use_case_id;type:varchar(36)"`
	Title           string     `gorm:"column:title;type:varchar(255)"`
	Description     string     `gorm:"column:description;type:text"`
	Active          *bool      `gorm:"column:active;type:bool"`
	Fallback        *bool      `gorm:"column:fallback;type:bool"`
	CurrentServePct *float64   `gorm:"column:current_pct;type:double precision"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
	ClonedFromID    *uuid.UUID `gorm:"-"`
}

func (m flowModel) TableName() string {
	return "mm_flow"
}

func (m flowModel) toEntity() flowEntity {
	return flowEntity(m)
}

type flowOrderBy string

const (
	flowOrderByTitle     flowOrderBy = "title"
	flowOrderByActive    flowOrderBy = "active"
	flowOrderByCreatedAt flowOrderBy = "created_at"
	flowOrderByUpdatedAt flowOrderBy = "updated_at"
	flowOrderByRelevance flowOrderBy = mm_db.RelevanceField
)

var availableFlowOrderBy = []interface{}{
	flowOrderByTitle,
	flowOrderByActive,
	flowOrderByCreatedAt,
	flowOrderByUpdatedAt,
	flowOrderByRelevance,
}
