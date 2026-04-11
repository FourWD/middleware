package infra

import (
	"context"
	"encoding/json"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type PubSubClient struct {
	client *gcppubsub.Client
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
	return &PubSubClient{client: client}, nil
}

func (c *PubSubClient) Publish(ctx context.Context, topic, message string) error {
	result := c.client.Topic(topic).Publish(ctx, &gcppubsub.Message{Data: []byte(message)})
	_, err := result.Get(ctx)
	return err
}

func (c *PubSubClient) PublishJSON(ctx context.Context, topic, prefix string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	return c.Publish(ctx, topic, prefix+"@"+string(b))
}

func (c *PubSubClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
