package common

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/spf13/viper"
)

func GoogleCreateSubscribe(topicName string) *pubsub.Subscription {
	projectID := viper.GetString("google_project_id")

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

	return sub
}
