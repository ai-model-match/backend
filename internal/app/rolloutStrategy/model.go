package rolloutStrategy

import (
	"encoding/json"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
)

type useCaseModel struct {
	ID uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
}

func (m useCaseModel) TableName() string {
	return "mm_use_case"
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

func (m rolloutStrategyModel) toEntity() rolloutStrategyEntity {
	// Remap the stored JSON config in the object configuration
	var config mm_pubsub.RSConfiguration
	if err := json.Unmarshal(m.Configuration, &config); err != nil {
		return rolloutStrategyEntity{}
	}
	return rolloutStrategyEntity{
		ID:            m.ID,
		UseCaseID:     m.UseCaseID,
		RolloutState:  m.RolloutState,
		Configuration: config,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func (m *rolloutStrategyModel) fromEntity(e rolloutStrategyEntity) error {
	// Convert the object configuration in JSON for saving
	if config, err := json.Marshal(e.Configuration); err != nil {
		return err
	} else {
		m.ID = e.ID
		m.UseCaseID = e.UseCaseID
		m.RolloutState = e.RolloutState
		m.Configuration = config
		m.CreatedAt = e.CreatedAt
		m.UpdatedAt = e.UpdatedAt
		return nil
	}
}
