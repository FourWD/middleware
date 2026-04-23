package infra

import (
	"context"
	"encoding/json"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PubSubClient struct {
	client    *gcppubsub.Client
	projectID string
}

func NewPubSubClient(ctx context.Context, cfg PubSubConfig) (*PubSubClient, error) {
	opts := []option.ClientOption{}
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	}

	client, err := gcppubsub.NewClient(ctx, cfg.ProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating pubsub client: %w", err)
	}
	return &PubSubClient{client: client, projectID: cfg.ProjectID}, nil
}

func (c *PubSubClient) Publish(ctx context.Context, topic, message string) error {
	_, err := c.PublishMessage(ctx, topic, []byte(message))
	return err
}

// PublishMessage publishes raw bytes and returns the server-assigned message ID.
func (c *PubSubClient) PublishMessage(ctx context.Context, topic string, data []byte) (string, error) {
	topicPath := fmt.Sprintf("projects/%s/topics/%s", c.projectID, topic)
	result := c.client.Publisher(topicPath).Publish(ctx, &gcppubsub.Message{Data: data})
	return result.Get(ctx)
}

func (c *PubSubClient) PublishJSON(ctx context.Context, topic, prefix string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	return c.Publish(ctx, topic, prefix+"@"+string(b))
}

// Subscribe blocks and invokes handler for each message delivered on the given
// subscription. The handler MUST call msg.Ack() or msg.Nack() to control
// redelivery. Returns when ctx is cancelled or an unrecoverable error occurs.
//
// This does NOT create the subscription — call EnsureSubscription at startup
// or provision it via Terraform/gcloud beforehand.
func (c *PubSubClient) Subscribe(
	ctx context.Context,
	subscriptionID string,
	handler func(ctx context.Context, msg *gcppubsub.Message),
) error {
	path := fmt.Sprintf("projects/%s/subscriptions/%s", c.projectID, subscriptionID)
	return c.client.Subscriber(path).Receive(ctx, handler)
}

// EnsureTopic creates the topic if missing. Idempotent — returns nil when the
// topic already exists.
func (c *PubSubClient) EnsureTopic(ctx context.Context, topicID string) error {
	topicPath := fmt.Sprintf("projects/%s/topics/%s", c.projectID, topicID)
	_, err := c.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicPath})
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("create topic %s: %w", topicID, err)
	}
	return nil
}

// EnsureSubscription creates the subscription if missing, linked to the given
// topic. Idempotent — returns nil when the subscription already exists.
func (c *PubSubClient) EnsureSubscription(ctx context.Context, topicID, subscriptionID string) error {
	topicPath := fmt.Sprintf("projects/%s/topics/%s", c.projectID, topicID)
	subPath := fmt.Sprintf("projects/%s/subscriptions/%s", c.projectID, subscriptionID)

	_, err := c.client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:  subPath,
		Topic: topicPath,
	})
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("create subscription %s: %w", subscriptionID, err)
	}
	return nil
}

func (c *PubSubClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
