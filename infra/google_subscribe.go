package infra

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

// GoogleSubscribe is a legacy wrapper that ensures a subscription named
// "SUB-<topicName>" exists, then blocks on Receive with a 2-second retry
// loop on error (matching the pre-refactor behaviour). Handler must NOT call
// msg.Ack() — this wrapper auto-acks after process returns, preserving the
// original contract.
//
// Prefer PubSub.Subscribe in new code — it surfaces errors and lets the
// caller control ack/nack + retry policy.
func GoogleSubscribe(topicName string, process func(message *pubsub.Message)) {
	if PubSub == nil {
		AppLog.EventWarn("PUBSUB_CLIENT_NOT_INITIALIZED", map[string]any{
			"topic": topicName,
		}, "",
			WithComponent(ComponentPubSub),
			WithOperation("subscribe"),
			WithLogKind(LogKindError))
		return
	}

	subscriptionID := "SUB-" + topicName

	AppLog.Event("PUBSUB_SUBSCRIBE_START", map[string]any{
		"topic":        topicName,
		"subscription": subscriptionID,
	}, "",
		WithComponent(ComponentPubSub),
		WithOperation("subscribe"),
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
		return
	}

	for {
		AppLog.Event("PUBSUB_LISTENING", map[string]any{"topic": topicName}, "",
			WithComponent(ComponentPubSub),
			WithOperation("listen"),
			WithLogKind(LogKindLifecycle))

		err := PubSub.Subscribe(ctx, subscriptionID, func(ctx context.Context, msg *pubsub.Message) {
			AppLog.Event("PUBSUB_MESSAGE_RECEIVED", map[string]any{
				"topic": topicName,
			}, "",
				WithComponent(ComponentPubSub),
				WithOperation("receive"),
				WithLogKind(LogKindBusiness))
			process(msg)
			msg.Ack()
		})

		if err != nil {
			LogPubSubError(ctx, err, "receive", topicName)
			time.Sleep(2 * time.Second)
		}

		if ctx.Err() != nil {
			AppLog.Event("PUBSUB_CONTEXT_CANCELLED", map[string]any{"topic": topicName}, "",
				WithComponent(ComponentPubSub),
				WithOperation("subscribe"),
				WithLogKind(LogKindLifecycle))
			return
		}
	}
}
