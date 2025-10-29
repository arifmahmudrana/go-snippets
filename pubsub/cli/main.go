package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/arifmahmudrana/go-snippets/pubsub"
)

func main() {
	// Create a new broker
	broker := pubsub.NewBroker()

	// Stop the broker after 10 seconds
	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("[BROKER] Stopping broker...")
		broker.Stop()
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	// Start Subscriber 1 (listens to "news")
	go func() {
		defer wg.Done()
		sub := broker.Subscribe("news")
		fmt.Println("[SUB 1] Subscribed to 'news'")

		for msg := range sub {
			fmt.Printf("[SUB 1] Received: %s\n", msg.Payload)
		}

		fmt.Println("[SUB 1] Unsubscribed (channel closed)")
	}()

	// Start Subscriber 2 (listens to "sports")
	go func() {
		defer wg.Done()
		sub := broker.Subscribe("sports")
		fmt.Println("[SUB 2] Subscribed to 'sports'")

		for msg := range sub {
			fmt.Printf("[SUB 2] Received: %s\n", msg.Payload)
		}

		fmt.Println("[SUB 2] Unsubscribed (channel closed)")
	}()

	// Give subscribers a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Start Publisher
	fmt.Println("[PUB] Publishing messages...")
	broker.Publish("news", "Heat wave expected next week")
	broker.Publish("sports", "Local team wins championship!")
	broker.Publish("news", "New technology announced")

	// Wait for subscribers to finish (which happens when broker stops)
	wg.Wait()
	fmt.Println("[MAIN] Application finished.")
}
