package mm_pubsub

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type eventModel struct {
	ID        uuid.UUID       `gorm:"primaryKey;column:id;type:varchar(36)"`
	Topic     string          `gorm:"column:topic;type:varchar(255)"`
	EventType string          `gorm:"column:event_type;type:varchar(255)"`
	EventDate time.Time       `gorm:"column:event_date;type:timestamp"`
	EventBody json.RawMessage `gorm:"column:event_body;type:json"`
}

func (m eventModel) TableName() string {
	return "mm_event"
}
