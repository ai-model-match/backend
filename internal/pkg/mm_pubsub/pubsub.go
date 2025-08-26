package mm_pubsub

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
PubSubMessage represents a generic message in pub-sub that is forwarded to consumers via channels.
It contains the Event with a pre-defined structured and the context of the call.
*/
type PubSubMessage struct {
	Message PubSubEvent
}

/*
PubSubAgent is a pub-sub agent that orchestrates channels to forward messages from producers to consumers.
*/
type PubSubAgent struct {
	mu     sync.Mutex
	subs   map[string][]chan PubSubMessage
	quit   chan struct{}
	closed bool
}

/*
NewPubSubAgent initialies a new pub-sub Agent.
*/
func NewPubSubAgent() *PubSubAgent {
	zap.L().Info("Start creatimg PubSub agent...", zap.String("service", "pub-sub"))
	pubsub := &PubSubAgent{
		subs: make(map[string][]chan PubSubMessage),
		quit: make(chan struct{}),
	}
	zap.L().Info("PubSub agent created!", zap.String("service", "pub-sub"))
	return pubsub
}

/*
Publish a message to a specific topic. The message will be sent to all the active channels.
*/
func (b *PubSubAgent) Publish(tx *gorm.DB, pubsubTopic PubSubTopic, msg PubSubMessage) error {
	if err := b.storeMessage(tx, pubsubTopic, msg); err != nil {
		return err
	}
	go b.publishMessageToTopic(pubsubTopic, msg)
	return nil
}

/*
Persist the new message on DB
*/
func (b *PubSubAgent) storeMessage(tx *gorm.DB, pubsubTopic PubSubTopic, msg PubSubMessage) error {
	rawMessage, err := json.Marshal(msg.Message)
	if err != nil {
		return err
	}
	model := eventModel{
		ID:        msg.Message.EventID,
		Topic:     string(pubsubTopic),
		EventType: string(msg.Message.EventType),
		EventDate: msg.Message.EventTime,
		EventBody: rawMessage,
	}
	return tx.Create(model).Error
}

/*
Publish a message to a specific topic. The message will be sent to all the active channels.
*/
func (b *PubSubAgent) publishMessageToTopic(pubsubTopic PubSubTopic, msg PubSubMessage) {
	topic := string(pubsubTopic)
	zap.L().Info(
		fmt.Sprintf("Dispatching %s event on Topic %s", msg.Message.EventType, topic),
		zap.String("service", "pub-sub"),
		zap.String("event", string(msg.Message.EventType)),
		zap.String("topic", topic),
	)
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	for _, ch := range b.subs[topic] {
		ch <- msg
	}
}

/*
Replay historical events optionally filtered by topic and start date
*/
func (b *PubSubAgent) ReplayMessages(tx *gorm.DB, topicName *PubSubTopic, startFromTime *time.Time) error {
	// Create query
	query := tx.Model(eventModel{})
	if topicName != nil {
		query.Where("topic = ?", topicName)
	}
	if startFromTime != nil {
		query.Where("event_date >= ?", startFromTime)
	}
	query.Order("event_date ASC")
	rows, err := query.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	// Process rows with SCAN approach to limit memory
	for rows.Next() {
		// Read the row
		var model eventModel
		if err := tx.ScanRows(rows, &model); err != nil {
			return err
		}
		// Unmarhal the stored event
		var body PubSubEvent
		if err := json.Unmarshal(model.EventBody, &body); err != nil {
			return err
		}
		// Convert EventEntity to raw bytes for further unmarshaling
		entityBytes, err := json.Marshal(body.EventEntity)
		if err != nil {
			return err
		}
		// Use factory to get typed struct
		factory, ok := eventEntityFactories[body.EventType]
		if !ok {
			return fmt.Errorf("unsupported event type: %s", body.EventType)
		}
		entityPtr := factory()
		if err := json.Unmarshal(entityBytes, entityPtr); err != nil {
			return err
		}
		// Recreate new typed body
		newBody := PubSubEvent{
			EventID:     body.EventID,
			EventTime:   body.EventTime,
			EventType:   body.EventType,
			EventEntity: entityPtr,
		}
		message := PubSubMessage{
			Message: newBody,
		}
		// Resend the event, without re-storing it, in SYNCHRONOUS way
		b.publishMessageToTopic(PubSubTopic(model.Topic), message)
	}
	return nil
}

/*
Subscribe to a topic by receving a dedicated channel to listen and wait published messages.
*/
func (b *PubSubAgent) Subscribe(pubsubTopic PubSubTopic) <-chan PubSubMessage {
	topic := string(pubsubTopic)
	zap.L().Info(
		fmt.Sprintf("Subscribing to Topic %s", topic),
		zap.String("service", "pub-sub"),
		zap.String("topic", topic),
	)
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	ch := make(chan PubSubMessage, 1)
	b.subs[topic] = append(b.subs[topic], ch)
	return ch
}

/*
Close the agent and all the channel avoiding publishers and consumers to send and read new events.
*/
func (b *PubSubAgent) Close() {
	zap.L().Info("Closing PubSub agent...", zap.String("service", "pub-sub"))
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	close(b.quit)

	for _, ch := range b.subs {
		for _, sub := range ch {
			close(sub)
		}
	}
	zap.L().Info("PubSub agent closed!", zap.String("service", "pub-sub"))
}
