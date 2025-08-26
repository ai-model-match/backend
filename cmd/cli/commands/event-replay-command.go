package commands

import (
	"errors"
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/urfave/cli"
	"gorm.io/gorm"
)

/*
EventReplayCommand is just an example of a command you can build and integtrate withint the CLI.
*/
func EventReplayCommand(c *cli.Context, m *mm_pubsub.PubSubAgent, tx *gorm.DB) error {
	startFrom := c.String("start-from")
	topicName := c.String("topic-name")

	// Validate start-from if set
	var startFromTime *time.Time
	var topic *mm_pubsub.PubSubTopic
	if startFrom != "" {
		if fromTime, err := time.Parse(time.RFC3339, startFrom); err != nil {
			return errors.New("start-from must be a valid ISO 8601 date, e.g., 2025-08-26T15:04:05Z")
		} else {
			startFromTime = &fromTime
		}
	}
	// Validate topic-name if set
	if topicName != "" {
		topic = (*mm_pubsub.PubSubTopic)(&topicName)
	}
	// Execute the command
	if err := m.ReplayMessages(tx, topic, startFromTime); err != nil {
		return err
	}
	return nil
}
