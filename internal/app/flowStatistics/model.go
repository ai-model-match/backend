package flowStatistics

import (
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
func (m flowModel) toEntity() flowEntity {
	return flowEntity(m)
}

type flowStatisticsModel struct {
	ID                 uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	FlowID             uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
	UseCaseID          uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	TotRequests        *int64    `gorm:"column:tot_req;type:bigint"`
	TotSessionRequests *int64    `gorm:"column:tot_sess_req;type:bigint"`
	InitialServePct    *float64  `gorm:"column:initial_pct;type:double precision"`
	CurrentServePct    *float64  `gorm:"column:current_pct;type:double precision"`
	TotFeedback        *int64    `gorm:"column:tot_feedback;type:bigint"`
	AvgFeedbackScore   *float64  `gorm:"column:avg_feedback_score;type:double precision"`
	CreatedAt          time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt          time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m flowStatisticsModel) TableName() string {
	return "mm_flow_statistics"
}

func (m flowStatisticsModel) toEntity() flowStatisticsEntity {
	return flowStatisticsEntity(m)
}
