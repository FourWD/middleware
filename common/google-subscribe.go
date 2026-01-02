package common

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

func GoogleSubscribe(topicName string, process func(message *pubsub.Message)) {
	Log("PUBSUB_SUBSCRIBE_START", map[string]interface{}{"topic": topicName}, "")

	pubsubCtx, err := initPubSubClient(topicName)
	if err != nil {
		return
	}
	defer pubsubCtx.Client.Close()

	subscriber := pubsubCtx.Client.Subscriber(pubsubCtx.SubscriptionPath)

	for {
		Log("PUBSUB_LISTENING", map[string]interface{}{"topic": topicName}, "")
		err := subscriber.Receive(pubsubCtx.Ctx, func(ctx context.Context, msg *pubsub.Message) {
			Log("PUBSUB_MESSAGE_RECEIVED", map[string]interface{}{"topic": topicName, "data": string(msg.Data)}, "")
			process(msg)
			msg.Ack()
		})

		if err != nil {
			LogError("PUBSUB_RECEIVE_ERROR", map[string]interface{}{"error": err.Error(), "topic": topicName}, "")
			LogWarning("PUBSUB_RETRYING", map[string]interface{}{"topic": topicName}, "")
			time.Sleep(2 * time.Second)
		}

		if pubsubCtx.Ctx.Err() != nil {
			Log("PUBSUB_CONTEXT_CANCELLED", map[string]interface{}{"topic": topicName}, "")
			return
		}
	}
}
