package infra

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
)

// GoogleCreateSubscribe is a legacy wrapper that ensures a subscription named
// "SUB-<topicName>" exists for the given topic and returns a Subscriber bound
// to it. Returns nil when the managed PubSub client is not initialized or
// subscription provisioning fails.
//
// Prefer PubSub.EnsureSubscription + PubSub.Subscribe in new code.
func GoogleCreateSubscribe(topicName string) *pubsub.Subscriber {
	if PubSub == nil {
		AppLog.EventWarn("PUBSUB_CLIENT_NOT_INITIALIZED", map[string]interface{}{
			"topic": topicName,
		}, "")
		return nil
	}

	subscriptionID := "SUB-" + topicName

	AppLog.Event("PUBSUB_CREATE_SUBSCRIBE", map[string]interface{}{
		"topic":        topicName,
		"subscription": subscriptionID,
	}, "")

	ctx := context.Background()
	if err := PubSub.EnsureSubscription(ctx, topicName, subscriptionID); err != nil {
		AppLog.EventError(err, "PUBSUB_SUBSCRIPTION_ENSURE_ERROR", map[string]interface{}{
			"topic":        topicName,
			"subscription": subscriptionID,
		}, "")
		return nil
	}

	subPath := fmt.Sprintf("projects/%s/subscriptions/%s", PubSub.projectID, subscriptionID)
	return PubSub.client.Subscriber(subPath)
}
