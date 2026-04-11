package common

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/spf13/viper"
)

type PubSubContext struct {
	Client           *pubsub.Client
	Ctx              context.Context
	TopicPath        string
	SubscriptionPath string
	SubscriptionID   string
}

func initPubSubClient(topicName string) (*PubSubContext, error) {
	projectID := viper.GetString("google_project_id")
	ctx := context.Background()
	subscriptionID := "SUB-" + topicName

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		LogError("PUBSUB_CLIENT_ERROR", map[string]interface{}{"error": err.Error(), "topic": topicName}, "")
		return nil, err
	}

	topicPath := fmt.Sprintf("projects/%s/topics/%s", projectID, topicName)
	subscriptionPath := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriptionID)

	_, err = client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{
		Topic: topicPath,
	})
	if err != nil {
		LogWarning("PUBSUB_TOPIC_NOT_EXISTS", map[string]interface{}{"topic": topicName, "error": err.Error()}, "")
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
			LogError("PUBSUB_SUBSCRIPTION_CREATE_ERROR", map[string]interface{}{"error": err.Error(), "subscription": subscriptionID}, "")
			client.Close()
			return nil, err
		}
		Log("PUBSUB_SUBSCRIPTION_CREATED", map[string]interface{}{"subscription": subscriptionID}, "")
	} else {
		Log("PUBSUB_SUBSCRIPTION_EXISTS", map[string]interface{}{"subscription": subscriptionID}, "")
	}

	return &PubSubContext{
		Client:           client,
		Ctx:              ctx,
		TopicPath:        topicPath,
		SubscriptionPath: subscriptionPath,
		SubscriptionID:   subscriptionID,
	}, nil
}
