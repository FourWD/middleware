package common

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
)

// func Subscribe(projectID, topicName string) {
func GoogleSubscribe(projectID string, topicName string, process func(topic string, message *pubsub.Message)) {
	log.Println("start", topicName)
	ctx := context.Background()
	subscriptionID := "SUB-" + topicName

	// Initialize Pub/Sub client once
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Error creating Pub/Sub client: %v", err)
	}
	defer client.Close()

	// Get or create the topic once
	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Printf("Error checking if topic exists: %v", err)
	}
	if !exists {
		log.Printf("Topic %s does not exist", topicName)
	}

	// Get or create the subscription once
	sub := client.Subscription(subscriptionID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		log.Printf("Error checking if subscription exists: %v", err)
	}
	if !exists {
		// Create the subscription if it doesn't exist
		sub, err = client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			log.Printf("Failed to create subscription: %v", err)
		}
		log.Printf("Created subscription: %s", subscriptionID)
	} else {
		log.Printf("Subscription already exists: %s", subscriptionID)
	}

	// Keep listening for messages, retrying only if an error occurs
	for {
		log.Printf("Listening for messages on topic: %s", topicName)
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			log.Println("sub.Receive", topicName, string(msg.Data))
			// routerUtils.SubManage(topicName, msg)
			process(topicName, msg)
			msg.Ack() // Acknowledge message after processing
		})

		// Handle error and retry
		if err != nil {
			log.Printf("Error receiving messages: %v", err)
			log.Println("Retrying message reception after error...")
			time.Sleep(2 * time.Second) // Add a delay to avoid tight retry loops
		}

		// Check if context is cancelled to avoid infinite retries
		if ctx.Err() != nil {
			log.Println("Context cancelled, stopping subscription")
			return
		}
	}
}
