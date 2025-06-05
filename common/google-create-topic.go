package common

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/spf13/viper"
)

func GoogleCreateTopic(topic string) error {
	projectID := viper.GetString("google_project_id")

	logFields := map[string]interface{}{
		"project_id": projectID,
		"topic":      topic,
	}
	Log("GoogleCreateTopic", logFields, "")

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.CreateTopic(ctx, topic)
	if err != nil {
		return err
	}

	return nil
}
