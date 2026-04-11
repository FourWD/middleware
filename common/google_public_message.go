package common

import (
	"context"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/FourWD/middleware/model"
	"github.com/spf13/viper"
)

func GooglePublicMessage(topicName, message string) (string, error) {
	projectID := viper.GetString("google_project_id")

	logData := map[string]interface{}{
		"project_id": projectID,
		"topic_name": topicName,
		"message":    message,
	}
	Log("GooglePublicMessage", logData, "")

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

func ConventStringToGoogleMessage(input string) model.GoogleMessage {
	gMessage := new(model.GoogleMessage)
	parts := strings.SplitN(input, "@", 2)

	if len(parts) == 2 {
		gMessage.Group = parts[0]
		gMessage.Message = parts[1]
	} else {
		gMessage.Group = ""
		gMessage.Message = input
	}

	return *gMessage
}
