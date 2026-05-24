package infra

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
)

// firebaseCallTimeout bounds every Firebase Messaging API call so a slow or
// unreachable Firebase backend cannot hang the caller's goroutine. Mirrors
// the value used by common/firebase_message.go.
const firebaseCallTimeout = 30 * time.Second

// MessageConfig holds the default Android/APNS delivery config used by
// outbound Firebase messages. Exported so callers can override before send.
var MessageConfig = struct {
	AndroidConfig *messaging.AndroidConfig
	APNSConfig    *messaging.APNSConfig
}{
	AndroidConfig: &messaging.AndroidConfig{
		Priority: "high",
	},
	APNSConfig: &messaging.APNSConfig{
		Headers: map[string]string{"apns-priority": "10"},
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Sound: "default",
			},
		},
	},
}

// SendMessageToUser sends a Firebase message to a single device token using
// MessageConfig for Android/APNS defaults.
func SendMessageToUser(userToken string, title string, body string, data map[string]string) error {
	if userToken == "" {
		return fmt.Errorf("user token is empty")
	}

	message := &messaging.Message{
		Data:  data,
		Token: userToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: MessageConfig.AndroidConfig,
		APNS:    MessageConfig.APNSConfig,
	}

	ctx, cancel := context.WithTimeout(context.Background(), firebaseCallTimeout)
	defer cancel()

	if _, err := FirebaseMessageClient.Send(ctx, message); err != nil {
		AppLog.EventError(err, "FIREBASE_MESSAGE_SEND_FAILURE", map[string]any{
			"title": title,
		}, "",
			WithComponent(ComponentFirebase),
			WithOperation("send_message"),
			WithLogKind(LogKindError))
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

// SendMessageToSubscriber broadcasts a Firebase message to every device
// subscribed to the given topic.
func SendMessageToSubscriber(topic string, title string, body string, data map[string]string) error {
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}

	requestID := uuid.NewString()
	logData := map[string]any{
		"topic": topic,
		"title": title,
	}
	AppLog.Event("FIREBASE_BROADCAST_START", logData, requestID,
		WithComponent(ComponentFirebase),
		WithOperation("broadcast"),
		WithLogKind(LogKindBusiness))

	message := &messaging.Message{
		Data:  data,
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), firebaseCallTimeout)
	defer cancel()

	if _, err := FirebaseMessageClient.Send(ctx, message); err != nil {
		AppLog.EventError(err, "FIREBASE_BROADCAST_FAILURE", logData, requestID,
			WithComponent(ComponentFirebase),
			WithOperation("broadcast"),
			WithLogKind(LogKindError))
		return fmt.Errorf("firebase broadcast: %w", err)
	}
	return nil
}
