package rolloutStrategy

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type useCaseModel struct {
	ID uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
}

func (m useCaseModel) TableName() string {
	return "mm_use_case"
}

type RolloutState string

const (
	RolloutStateInit            RolloutState = "INIT"
	RolloutStateWarmup          RolloutState = "WARMUP"
	RolloutStateEscaped         RolloutState = "ESCAPED"
	RolloutStateMonitor         RolloutState = "MONITOR"
	RolloutStateAdaptive        RolloutState = "ADAPTIVE"
	RolloutStateCompleted       RolloutState = "COMPLETED"
	RolloutStateForcedEscaped   RolloutState = "FORCED_ESCAPED"
	RolloutStateForcedCompleted RolloutState = "FORCED_COMPLETED"
)

type rolloutStrategyModel struct {
	ID            uuid.UUID       `gorm:"primaryKey;column:id;type:varchar(36)"`
	UseCaseID     uuid.UUID       `gorm:"column:use_case_id;type:varchar(36)"`
	RolloutState  RolloutState    `gorm:"column:rollout_state;type:rollout_state"`
	Configuration json.RawMessage `gorm:"column:configuration;type:json"`
	CreatedAt     time.Time       `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
}

func (m rolloutStrategyModel) TableName() string {
	return "mm_rollout_strategy"
}

func (m rolloutStrategyModel) toEntity() rolloutStrategyEntity {
	return rolloutStrategyEntity(m)
}

var AvailableRolloutState = []interface{}{
	RolloutStateInit,
	RolloutStateWarmup,
	RolloutStateEscaped,
	RolloutStateMonitor,
	RolloutStateAdaptive,
	RolloutStateCompleted,
	RolloutStateForcedEscaped,
	RolloutStateForcedCompleted,
}

var allowedTransitions = map[RolloutState][]RolloutState{
	RolloutStateInit:            {RolloutStateWarmup},
	RolloutStateWarmup:          {RolloutStateForcedEscaped, RolloutStateForcedCompleted},
	RolloutStateMonitor:         {RolloutStateForcedEscaped, RolloutStateForcedCompleted},
	RolloutStateAdaptive:        {RolloutStateForcedEscaped, RolloutStateForcedCompleted},
	RolloutStateCompleted:       {RolloutStateInit},
	RolloutStateForcedEscaped:   {RolloutStateInit},
	RolloutStateForcedCompleted: {RolloutStateInit},
}
