package flowStep

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type flowModel struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
}

func (m flowModel) TableName() string {
	return "mm_flow"
}

type useCaseStepModel struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	Position  int64     `gorm:"column:position;type:bigint"`
}

func (m useCaseStepModel) TableName() string {
	return "mm_use_case_step"
}

type flowStepModel struct {
	ID            uuid.UUID       `gorm:"primaryKey;column:id;type:varchar(36)"`
	FlowID        uuid.UUID       `gorm:"column:flow_id;type:varchar(36)"`
	UseCaseID     uuid.UUID       `gorm:"column:use_case_id;type:varchar(36)"`
	UseCaseStepID uuid.UUID       `gorm:"column:use_case_step_id;type:varchar(36)"`
	Configuration json.RawMessage `gorm:"column:configuration;type:json"`
	Placeholders  json.RawMessage `gorm:"column:placeholders;type:json"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m flowStepModel) TableName() string {
	return "mm_flow_step"
}

func (m flowStepModel) toEntity() flowStepEntity {
	return flowStepEntity(m)
}

type missingFlowStepModel struct {
	FlowID        uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
	UseCaseID     uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	UseCaseStepID uuid.UUID `gorm:"column:use_case_step_id;type:varchar(36)"`
}

func (m missingFlowStepModel) toEntity() missingFlowStepEntity {
	return missingFlowStepEntity(m)
}
