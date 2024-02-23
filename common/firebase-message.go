package common

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var FirebaseMessageClient *messaging.Client

func ConnectFirebaseNotification(key string) error {
	opt := option.WithCredentialsFile(key)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	FirebaseMessageClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
		return err
	}

	return nil
}

func SendNotificationToUser(userToken string, data map[string]string) error {
	message := &messaging.Message{
		Data:  data,
		Token: userToken,
	}

	_, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		log.Fatalf("error sending message: %v\n", err)
		return err
	}

	return nil
}

func AddNotificationSubscribe(userToken string, topic string) error {
	_, err := FirebaseMessageClient.SubscribeToTopic(context.Background(), []string{userToken}, topic)
	if err != nil {
		log.Fatalf("error subscribing user to topic: %v\n", err)
		return err
	}
	return nil
}

func RemoveNotificationSubscribe(userToken string, topic string) error {
	_, err := FirebaseMessageClient.UnsubscribeFromTopic(context.Background(), []string{userToken}, topic)
	if err != nil {
		log.Fatalf("error unsubscribing user from topic: %v\n", err)
		return err
	}
	return nil
}

func SendNotificationToSubscribe(userToken string, topic string, data map[string]string) error {
	message := &messaging.Message{
		Data:  data,
		Topic: topic, // all_users
	}

	_, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		log.Fatalf("error sending message: %v\n", err)
		return err
	}

	return nil
}

// message := &messaging.Message{
// 	Notification: &messaging.Notification{
// 		Title: "Title of Notification",
// 		Body:  "Body of Notification",
// 	},
// 	Token: "device_token_or_topic",
// }
