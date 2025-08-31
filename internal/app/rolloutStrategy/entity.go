package rolloutStrategy

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type rolloutStrategyEntity struct {
	ID            uuid.UUID       `json:"id"`
	UseCaseID     uuid.UUID       `json:"useCaseId"`
	RolloutState  RolloutState    `json:"rolloutState"`
	Configuration json.RawMessage `json:"configuration"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}
