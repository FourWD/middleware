package common

import (
	"cloud.google.com/go/pubsub/v2"
	"github.com/spf13/viper"
)

func GoogleCreateSubscribe(topicName string) *pubsub.Subscriber {
	projectID := viper.GetString("google_project_id")

	logData := map[string]interface{}{
		"project_id": projectID,
		"topic":      topicName,
	}
	Log("PUBSUB_CREATE_SUBSCRIBE", logData, "")

	pubsubCtx, err := initPubSubClient(topicName)
	if err != nil {
		return nil
	}
	defer pubsubCtx.Client.Close()

	return pubsubCtx.Client.Subscriber(pubsubCtx.SubscriptionPath)
}
