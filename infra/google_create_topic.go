package infra

import (
	"context"
	"errors"
)

// GoogleCreateTopic is a legacy wrapper that creates a Pub/Sub topic if it
// does not exist, via the managed PubSub client. Prefer PubSub.EnsureTopic
// in new code — it returns nil when the topic already exists and is safe to
// call idempotently at startup.
func GoogleCreateTopic(topic string) error {
	if PubSub == nil {
		return errors.New("pubsub client not initialized; set PUBSUB_ENABLED=true")
	}

	AppLog.Event("GoogleCreateTopic", map[string]interface{}{
		"topic": topic,
	}, "")

	return PubSub.EnsureTopic(context.Background(), topic)
}
