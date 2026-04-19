package infra

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
)

type PubSubContext struct {
	Client           *pubsub.Client
	Ctx              context.Context
	TopicPath        string
	SubscriptionPath string
	SubscriptionID   string
}

func initPubSubClient(topicName string) (*PubSubContext, error) {
	projectID := GetEnv("GCP_PROJECT_ID", "")
	ctx := context.Background()
	subscriptionID := "SUB-" + topicName

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		AppLog.EventError(err, "PUBSUB_CLIENT_ERROR", map[string]interface{}{"topic": topicName}, "")
		return nil, err
	}

	topicPath := fmt.Sprintf("projects/%s/topics/%s", projectID, topicName)
	subscriptionPath := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriptionID)

	_, err = client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{
		Topic: topicPath,
	})
	if err != nil {
		AppLog.EventWarn("PUBSUB_TOPIC_NOT_EXISTS", map[string]interface{}{"topic": topicName, "error": err.Error()}, "")
	}

	_, err = client.SubscriptionAdminClient.GetSubscription(ctx, &pubsubpb.GetSubscriptionRequest{
		Subscription: subscriptionPath,
	})
	if err != nil {
		_, err = client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
			Name:  subscriptionPath,
			Topic: topicPath,
		})
		if err != nil {
			AppLog.EventError(err, "PUBSUB_SUBSCRIPTION_CREATE_ERROR", map[string]interface{}{"subscription": subscriptionID}, "")
			client.Close()
			return nil, err
		}
		AppLog.Event("PUBSUB_SUBSCRIPTION_CREATED", map[string]interface{}{"subscription": subscriptionID}, "")
	} else {
		AppLog.Event("PUBSUB_SUBSCRIPTION_EXISTS", map[string]interface{}{"subscription": subscriptionID}, "")
	}

	return &PubSubContext{
		Client:           client,
		Ctx:              ctx,
		TopicPath:        topicPath,
		SubscriptionPath: subscriptionPath,
		SubscriptionID:   subscriptionID,
	}, nil
}
