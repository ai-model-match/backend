package mmpubsub

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

/*
PubSubMessage represents a generic message in pub-sub that is forwarded to consumers via channels.
It contains the Event with a pre-defined structured and the context of the call.
*/
type PubSubMessage struct {
	Message PubSubEvent
	Context *gin.Context
}

/*
PubSubAgent is a pub-sub agent that orchestrates channels to forward messages from producers to consumers.
*/
type PubSubAgent struct {
	logger *zap.Logger
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
func (b *PubSubAgent) Publish(pubsubTopic PubSubTopic, msg PubSubMessage) {
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
