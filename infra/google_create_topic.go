package infra

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
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

	topicPath := fmt.Sprintf("projects/%s/topics/%s", projectID, topic)
	_, err = client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: topicPath,
	})
	if err != nil {
		return err
	}

	return nil
}
