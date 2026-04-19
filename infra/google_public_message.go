package infra

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub/v2"
	"github.com/FourWD/middleware/model"
)

func GooglePublicMessage(topicName, message string) (string, error) {
	projectID := GetEnv("GCP_PROJECT_ID", "")

	AppLog.Event("GooglePublicMessage", map[string]interface{}{
		"project_id": projectID,
		"topic_name": topicName,
		"message":    message,
	}, "")

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return "", err
	}
	defer client.Close()

	topicPath := fmt.Sprintf("projects/%s/topics/%s", projectID, topicName)
	result := client.Publisher(topicPath).Publish(ctx, &pubsub.Message{
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
