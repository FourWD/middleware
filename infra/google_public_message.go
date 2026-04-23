package infra

import (
	"context"
	"errors"
	"strings"

	"github.com/FourWD/middleware/model"
)

// GooglePublicMessage is a legacy wrapper that publishes a raw string message
// to the given Pub/Sub topic via the managed PubSub client (deps.Cloud.PubSub).
// Prefer calling PubSub.Publish(ctx, ...) directly in new code.
func GooglePublicMessage(topicName, message string) (string, error) {
	if PubSub == nil {
		return "", errors.New("pubsub client not initialized; set PUBSUB_ENABLED=true")
	}

	AppLog.Event("GooglePublicMessage", map[string]interface{}{
		"topic_name": topicName,
		"message":    message,
	}, "")

	ctx := context.Background()
	return PubSub.PublishMessage(ctx, topicName, []byte(message))
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
