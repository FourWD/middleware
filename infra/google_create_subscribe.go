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
		AppLog.EventWarn("PUBSUB_CLIENT_NOT_INITIALIZED", map[string]any{
			"topic": topicName,
		}, "",
			WithComponent(ComponentPubSub),
			WithOperation("subscribe"),
			WithLogKind(LogKindError))
		return nil
	}

	subscriptionID := "SUB-" + topicName

	AppLog.Event("PUBSUB_SUBSCRIPTION_ENSURE_START", map[string]any{
		"topic":        topicName,
		"subscription": subscriptionID,
	}, "",
		WithComponent(ComponentPubSub),
		WithOperation("ensure_subscription"),
		WithLogKind(LogKindBusiness))

	ctx := context.Background()
	if err := PubSub.EnsureSubscription(ctx, topicName, subscriptionID); err != nil {
		AppLog.EventError(err, "PUBSUB_SUBSCRIPTION_ENSURE_FAILURE", map[string]any{
			"topic":        topicName,
			"subscription": subscriptionID,
		}, "",
			WithComponent(ComponentPubSub),
			WithOperation("ensure_subscription"),
			WithLogKind(LogKindError))
		return nil
	}

	subPath := fmt.Sprintf("projects/%s/subscriptions/%s", PubSub.projectID, subscriptionID)
	return PubSub.client.Subscriber(subPath)
}
