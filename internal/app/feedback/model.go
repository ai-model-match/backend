package feedback

import (
	"time"

	"github.com/google/uuid"
)

type pickerCorrelationModel struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID uuid.UUID `gorm:"primaryKey;column:use_case_id;type:varchar(36)"`
	FlowID    uuid.UUID `gorm:"primaryKey;column:flow_id;type:varchar(36)"`
	Fallback  bool      `gorm:"column:fallback;type:bool"`
}

func (m pickerCorrelationModel) TableName() string {
	return "mm_picker_correlation"
}

func (m pickerCorrelationModel) toEntity() pickerCorrelationEntity {
	return pickerCorrelationEntity(m)
}

type feedbackModel struct {
	ID            uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID     uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	FlowID        uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
	CorrelationID uuid.UUID `gorm:"column:correlation_id;type:varchar(36)"`
	Score         float64   `gorm:"column:score;type:double precision"`
	Comment       string    `gorm:"column:comment;type:text"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
}

func (m feedbackModel) TableName() string {
	return "mm_feedback"
}
