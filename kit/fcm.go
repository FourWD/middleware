package kit

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

func DefaultAndroidConfig() *messaging.AndroidConfig {
	return &messaging.AndroidConfig{
		Priority: "high",
	}
}

func DefaultAPNSConfig() *messaging.APNSConfig {
	return &messaging.APNSConfig{
		Headers: map[string]string{"apns-priority": "10"},
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{Sound: "default"},
		},
	}
}

type FCMClient struct {
	client *messaging.Client
}

func NewFCMClient(ctx context.Context, credentialsFile string) (*FCMClient, error) {
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("creating firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating messaging client: %w", err)
	}

	return &FCMClient{client: msgClient}, nil
}

func (fc *FCMClient) SendToToken(ctx context.Context, token, title, body string, data map[string]string) (string, error) {
	msg := &messaging.Message{
		Data:         data,
		Token:        token,
		Notification: &messaging.Notification{Title: title, Body: body},
		Android:      DefaultAndroidConfig(),
		APNS:         DefaultAPNSConfig(),
	}
	return fc.client.Send(ctx, msg)
}

func (fc *FCMClient) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) (string, error) {
	msg := &messaging.Message{
		Data:         data,
		Topic:        topic,
		Notification: &messaging.Notification{Title: title, Body: body},
	}
	return fc.client.Send(ctx, msg)
}

func (fc *FCMClient) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := fc.client.SubscribeToTopic(ctx, tokens, topic)
	return err
}

func (fc *FCMClient) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := fc.client.UnsubscribeFromTopic(ctx, tokens, topic)
	return err
}
