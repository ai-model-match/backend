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
	return rolloutStrategyEntity(m)
}

var allowedTransitions = map[mm_pubsub.RolloutState][]mm_pubsub.RolloutState{
	mm_pubsub.RolloutStateInit:            {mm_pubsub.RolloutStateWarmup},
	mm_pubsub.RolloutStateWarmup:          {mm_pubsub.RolloutStateForcedEscaped, mm_pubsub.RolloutStateForcedCompleted},
	mm_pubsub.RolloutStateMonitor:         {mm_pubsub.RolloutStateForcedEscaped, mm_pubsub.RolloutStateForcedCompleted},
	mm_pubsub.RolloutStateAdaptive:        {mm_pubsub.RolloutStateForcedEscaped, mm_pubsub.RolloutStateForcedCompleted},
	mm_pubsub.RolloutStateCompleted:       {mm_pubsub.RolloutStateInit},
	mm_pubsub.RolloutStateForcedEscaped:   {mm_pubsub.RolloutStateInit},
	mm_pubsub.RolloutStateForcedCompleted: {mm_pubsub.RolloutStateInit},
}
