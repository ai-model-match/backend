package rsEngine

import (
	"encoding/json"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
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
	ID              uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID       uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	CurrentServePct *float64  `gorm:"column:current_pct;type:double precision"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m flowModel) TableName() string {
	return "mm_flow"
}

type flowStatisticsModel struct {
	ID                 uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	FlowID             uuid.UUID `gorm:"column:flow_id;type:varchar(36)"`
	UseCaseID          uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	TotRequests        int64     `gorm:"column:tot_req;type:bigint"`
	TotSessionRequests int64     `gorm:"column:tot_sess_req;type:bigint"`
	CurrentServePct    float64   `gorm:"column:current_pct;type:double precision"`
	TotFeedback        int64     `gorm:"column:tot_feedback;type:bigint"`
	AvgScore           float64   `gorm:"column:avg_score;type:double precision"`
}

func (m flowStatisticsModel) TableName() string {
	return "mm_flow_statistics"
}

type rolloutStrategyModel struct {
	ID            uuid.UUID              `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID     uuid.UUID              `gorm:"column:use_case_id;type:varchar(36)"`
	RolloutState  mm_pubsub.RolloutState `gorm:"column:rollout_state;type:rollout_state"`
	Configuration json.RawMessage        `gorm:"column:configuration;type:json"`
	CreatedAt     time.Time              `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt     time.Time              `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m rolloutStrategyModel) TableName() string {
	return "mm_rollout_strategy"
}
