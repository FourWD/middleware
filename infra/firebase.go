package infra

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FirebaseClient groups optional Firestore and FCM clients.
type FirebaseClient struct {
	Firestore    *firestore.Client
	Notification *FirebaseMessagingClient
}

type FirebaseMessagingClient struct {
	client *messaging.Client
}

func NewFirebaseClient(ctx context.Context, cfg FirebaseConfig) (*FirebaseClient, error) {
	var fsClient *firestore.Client
	var err error

	if cfg.CredentialsFile != "" {
		fsClient, err = firestore.NewClient(ctx, firestore.DetectProjectID, option.WithCredentialsFile(cfg.CredentialsFile))
		if err != nil {
			return nil, fmt.Errorf("create firestore client: %w", err)
		}
	}

	var fcmClient *FirebaseMessagingClient
	if cfg.NotificationCredentialsFile != "" {
		opt := option.WithCredentialsFile(cfg.NotificationCredentialsFile)
		app, appErr := firebase.NewApp(ctx, nil, opt)
		if appErr != nil {
			if fsClient != nil {
				_ = fsClient.Close()
			}
			return nil, fmt.Errorf("create firebase app: %w", appErr)
		}
		msgClient, msgErr := app.Messaging(ctx)
		if msgErr != nil {
			if fsClient != nil {
				_ = fsClient.Close()
			}
			return nil, fmt.Errorf("create fcm client: %w", msgErr)
		}
		fcmClient = &FirebaseMessagingClient{client: msgClient}
	}

	if fsClient == nil && fcmClient == nil {
		return nil, nil
	}

	return &FirebaseClient{
		Firestore:    fsClient,
		Notification: fcmClient,
	}, nil
}

func (c *FirebaseClient) Close() error {
	if c == nil || c.Firestore == nil {
		return nil
	}
	return c.Firestore.Close()
}

func (fc *FirebaseMessagingClient) SendToToken(ctx context.Context, token, title, body string, data map[string]string) (string, error) {
	msg := &messaging.Message{
		Data:         data,
		Token:        token,
		Notification: &messaging.Notification{Title: title, Body: body},
		Android:      &messaging.AndroidConfig{Priority: "high"},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{"apns-priority": "10"},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{Sound: "default"},
			},
		},
	}
	return fc.client.Send(ctx, msg)
}

func (fc *FirebaseMessagingClient) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) (string, error) {
	msg := &messaging.Message{
		Data:         data,
		Topic:        topic,
		Notification: &messaging.Notification{Title: title, Body: body},
	}
	return fc.client.Send(ctx, msg)
}

func (fc *FirebaseMessagingClient) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := fc.client.SubscribeToTopic(ctx, tokens, topic)
	return err
}

func (fc *FirebaseMessagingClient) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := fc.client.UnsubscribeFromTopic(ctx, tokens, topic)
	return err
}
