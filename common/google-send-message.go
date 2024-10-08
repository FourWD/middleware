package common

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/spf13/viper"
)

func GooglePublicMessage(topicName, message string) (string, error) {
	projectID := viper.GetString("google_project_id")

	log.Println("Message", topicName, message)
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return "", err
	}
	defer client.Close()

	topic := client.Topic(topicName)
	if topic == nil {
		return "", err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(message),
	})

	// Block until the result is returned
	id, err := result.Get(ctx)
	if err != nil {
		return "", err
	}

	return id, nil
}
