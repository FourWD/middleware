package infra

import (
	"context"

	"cloud.google.com/go/pubsub"
)

func GoogleCreateTopic(topic string) error {
	projectID := GetEnv("GCP_PROJECT_ID", "")

	AppLog.Event("GoogleCreateTopic", map[string]interface{}{
		"project_id": projectID,
		"topic":      topic,
	}, "")

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
