package rsEngine

import (
	"encoding/json"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type flowModel struct {
	ID              uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID       uuid.UUID `gorm:"column:use_case_id;type:varchar(36)"`
	Active          bool      `gorm:"column:active;type:bool"`
	Fallback        bool      `gorm:"column:fallback;type:bool"`
	CurrentServePct float64   `gorm:"column:current_pct;type:double precision"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
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
	TotRequests        int64     `gorm:"column:tot_req;type:bigint"`
	TotSessionRequests int64     `gorm:"column:tot_sess_req;type:bigint"`
	CurrentServePct    float64   `gorm:"column:current_pct;type:double precision"`
	TotFeedback        int64     `gorm:"column:tot_feedback;type:bigint"`
	AvgScore           float64   `gorm:"column:avg_score;type:double precision"`
}

func (m flowStatisticsModel) TableName() string {
	return "mm_flow_statistics"
}

func (m flowStatisticsModel) toEntity() flowStatisticsEntity {
	return flowStatisticsEntity(m)
}

type rolloutStrategyModel struct {
	ID            uuid.UUID              `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID     uuid.UUID              `gorm:"column:use_case_id;type:varchar(36)"`
	RolloutState  mm_pubsub.RolloutState `gorm:"column:rollout_state;type:rollout_state"`
	Configuration json.RawMessage        `gorm:"column:configuration;type:json"`
	UpdatedAt     time.Time              `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m rolloutStrategyModel) TableName() string {
	return "mm_rollout_strategy"
}

func (m rolloutStrategyModel) toEntity() rolloutStrategyEntity {
	var config mm_pubsub.RSConfiguration
	if err := json.Unmarshal(m.Configuration, &config); err != nil {
		return rolloutStrategyEntity{}
	}
	return rolloutStrategyEntity{
		ID:            m.ID,
		UseCaseID:     m.UseCaseID,
		RolloutState:  m.RolloutState,
		Configuration: config,
		UpdatedAt:     m.UpdatedAt,
	}
}

func (m *rolloutStrategyModel) fromEntity(e rolloutStrategyEntity) error {
	// Store only the necessary fields
	m.ID = e.ID
	m.RolloutState = e.RolloutState
	m.UpdatedAt = e.UpdatedAt
	return nil
}
