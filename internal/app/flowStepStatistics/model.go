package flowStepStatistics

import (
	"time"

	"github.com/google/uuid"
)

type flowStepModel struct {
	ID     uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	FlowID uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
}

func (m flowStepModel) TableName() string {
	return "mm_flow_step"
}
func (m flowStepModel) toEntity() flowStepEntity {
	return flowStepEntity(m)
}

type flowStepStatisticsModel struct {
	ID          uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	FlowStepID  uuid.UUID `gorm:"column:flow_step_id;type:varchar(36)"`
	FlowID      uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
	TotRequests *int64    `gorm:"column:tot_req;type:bigint"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m flowStepStatisticsModel) TableName() string {
	return "mm_flow_step_statistics"
}

func (m flowStepStatisticsModel) toEntity() flowStepStatisticsEntity {
	return flowStepStatisticsEntity(m)
}
