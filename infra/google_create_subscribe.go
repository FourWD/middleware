package infra

import (
	"cloud.google.com/go/pubsub/v2"
)

func GoogleCreateSubscribe(topicName string) *pubsub.Subscriber {
	projectID := GetEnv("GCP_PROJECT_ID", "")

	AppLog.Event("PUBSUB_CREATE_SUBSCRIBE", map[string]interface{}{
		"project_id": projectID,
		"topic":      topicName,
	}, "")

	pubsubCtx, err := initPubSubClient(topicName)
	if err != nil {
		return nil
	}
	defer pubsubCtx.Client.Close()

	return pubsubCtx.Client.Subscriber(pubsubCtx.SubscriptionPath)
}
