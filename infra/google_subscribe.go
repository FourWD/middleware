package infra

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

// GoogleSubscribe is a legacy wrapper that ensures a subscription named
// "SUB-<topicName>" exists, then blocks on Receive with a 2-second retry
// loop on error (matching the pre-refactor behaviour). The wrapper acks after
// process returns; process may call msg.Nack() to request redelivery (the
// first ack/nack wins). A panic is logged and the message nacked. Nacked
// messages redeliver, so process must be idempotent.
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
			if processRecovered(ctx, msg, process) {
				msg.Ack()
			}
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

// processRecovered runs process and reports whether it returned normally. On
// panic the message is nacked so Pub/Sub redelivers it.
func processRecovered(ctx context.Context, msg *pubsub.Message, process func(*pubsub.Message)) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			LogCritical(ctx, fmt.Errorf("panic: %v", r), "PUBSUB_HANDLER_PANIC", ComponentPubSub, "process_message")
			msg.Nack()
			ok = false
		}
	}()
	process(msg)
	return true
}
