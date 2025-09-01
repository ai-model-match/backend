package picker

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type useCaseModel struct {
	ID     uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	Code   string    `gorm:"column:code;type:varchar(255)"`
	Active bool      `gorm:"column:active;type:boolean"`
}

func (m useCaseModel) TableName() string {
	return "mm_use_case"
}

func (m useCaseModel) toEntity() useCaseEntity {
	return useCaseEntity(m)
}

type useCaseStepModel struct {
	ID   uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	Code string    `gorm:"column:code;type:varchar(255)"`
}

func (m useCaseStepModel) TableName() string {
	return "mm_use_case_step"
}

func (m useCaseStepModel) toEntity() useCaseStepEntity {
	return useCaseStepEntity(m)
}

type flowModel struct {
	ID              uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID       uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	Active          bool      `gorm:"column:active;type:bool"`
	Fallback        bool      `gorm:"column:fallback;type:bool"`
	CurrentServePct float64   `gorm:"column:current_pct;type:double precision"`
}

func (m flowModel) TableName() string {
	return "mm_flow"
}

func (m flowModel) toEntity() flowEntity {
	return flowEntity(m)
}

type flowStepModel struct {
	ID            uuid.UUID       `gorm:"primaryKey;column:id;type:varchar(36)"`
	FlowID        uuid.UUID       `gorm:"column:flow_id;type:varchar(36)"`
	UseCaseID     uuid.UUID       `gorm:"column:use_case_id;type:varchar(36)"`
	UseCaseStepID uuid.UUID       `gorm:"column:use_case_step_id;type:varchar(36)"`
	Configuration json.RawMessage `gorm:"column:configuration;type:json"`
	Placeholders  json.RawMessage `gorm:"column:placeholders;type:json"`
}

func (m flowStepModel) TableName() string {
	return "mm_flow_step"
}

func (m flowStepModel) toEntity() flowStepEntity {
	return flowStepEntity(m)
}

type pickerCorrelationModel struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	FlowID    uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
	Fallback  bool      `gorm:"column:fallback;type:bool"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
}

func (m pickerCorrelationModel) TableName() string {
	return "mm_picker_correlation"
}

func (m pickerCorrelationModel) toEntity() pickerCorrelationEntity {
	return pickerCorrelationEntity(m)
}
