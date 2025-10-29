package pubsub

import (
	"context"
	"time"
)

// Message holds the content being published.
type Message struct {
	Topic   string
	Payload interface{}
}

// Subscriber is a channel that receives messages.
// A subscriber client will read from this channel.
type Subscriber chan Message

// Broker is the central hub that manages topics, subscribers,
// and the broadcasting of messages.
type Broker struct {
	// A map of topics to a map of subscribers.
	// map[topic]map[subscriber]bool
	subscriptions map[string]map[Subscriber]bool

	// Channel for receiving new subscription requests.
	subCh chan subRequest

	// Channel for receiving unsubscription requests.
	unsubCh chan unsubRequest

	// Channel for receiving messages to be published.
	pubCh chan Message

	// Channel to signal the broker to stop.
	stopCh chan struct{}
}

// subRequest wraps a subscription request.
type subRequest struct {
	topic string
	sub   Subscriber
}

// unsubRequest wraps an unsubscription request.
type unsubRequest struct {
	topic string
	sub   Subscriber
}

// NewBroker creates and starts a new Broker.
func NewBroker() *Broker {
	b := &Broker{
		subscriptions: make(map[string]map[Subscriber]bool),
		subCh:         make(chan subRequest),
		unsubCh:       make(chan unsubRequest),
		pubCh:         make(chan Message),
		stopCh:        make(chan struct{}),
	}

	// Start the central run loop in a goroutine
	go b.run()
	return b
}

// run is the central loop that manages the broker's state.
// This is the *only* goroutine allowed to access the subscriptions map,
// which prevents data races.
func (b *Broker) run() {
	defer func() {
		// On exit, close all channels
		close(b.subCh)
		close(b.unsubCh)
		close(b.pubCh)
	}()

	for {
		select {
		case <-b.stopCh:
			// Signal to stop. Close all active subscriber channels.
			for _, topicSubs := range b.subscriptions {
				for sub := range topicSubs {
					close(sub)
				}
			}
			return

		case req := <-b.subCh:
			// New subscription
			if b.subscriptions[req.topic] == nil {
				b.subscriptions[req.topic] = make(map[Subscriber]bool)
			}
			b.subscriptions[req.topic][req.sub] = true

		case req := <-b.unsubCh:
			// Unsubscription
			if topicSubs, ok := b.subscriptions[req.topic]; ok {
				if _, subOk := topicSubs[req.sub]; subOk {
					// Delete the subscriber
					delete(topicSubs, req.sub)
					// Close its channel to signal it's been unsubscribed
					close(req.sub)
				}
			}

		case msg := <-b.pubCh:
			// New message published
			if topicSubs, ok := b.subscriptions[msg.Topic]; ok {
				// Broadcast to all subscribers of this topic
				for sub := range topicSubs {
					// Send the message in a new goroutine to prevent a slow
					// subscriber from blocking the entire broker.
					go func(s Subscriber, m Message) {
						// We can use a context with timeout to prevent
						// a non-reading goroutine from leaking forever.
						ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
						defer cancel()

						select {
						case s <- m:
						case <-ctx.Done():
							// Subscriber was too slow, message dropped.
						}
					}(sub, msg)
				}
			}
		}
	}
}

// Subscribe adds a new subscriber to a topic and returns the channel.
// We add a small buffer to the subscriber channel to reduce blocking.
func (b *Broker) Subscribe(topic string) Subscriber {
	sub := make(Subscriber, 10) // Buffered channel
	req := subRequest{
		topic: topic,
		sub:   sub,
	}

	b.subCh <- req
	return sub
}

// Unsubscribe removes a subscriber from a topic.
func (b *Broker) Unsubscribe(topic string, sub Subscriber) {
	req := unsubRequest{
		topic: topic,
		sub:   sub,
	}

	b.unsubCh <- req
}

// Publish broadcasts a message to all subscribers of a topic.
func (b *Broker) Publish(topic string, payload interface{}) {
	msg := Message{
		Topic:   topic,
		Payload: payload,
	}

	b.pubCh <- msg
}

// Stop shuts down the broker and closes all subscriber channels.
func (b *Broker) Stop() {
	close(b.stopCh)
}
