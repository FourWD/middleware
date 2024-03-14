package common

import (
	"context"
	"fmt"
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

func SendMessageToUser(userToken string, title string, body string, data map[string]string) error {
	// Access title and body directly from the data map

	message := &messaging.Message{
		Data:  data,
		Token: userToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	result, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		// Instead of log.Fatalf, return the error
		return fmt.Errorf("error sending message: %s, %v", result, err)
	}

	return nil
}

func AddUserToSubscription(userToken string, topic string) error { // เอาคน (topic) เข้า กรุป auction
	fmt.Printf("UserToken: %s, Topic: %s\n", userToken, topic)
	_, err := FirebaseMessageClient.SubscribeToTopic(context.Background(), []string{userToken}, topic)
	if err != nil {
		log.Fatalf("error subscribing user to topic: %v\n", err)
		return err
	}
	return nil
}

func RemoveUserFromSubscription(userToken string, topic string) error { // เอาคน (topic) ออก กรุป auction
	_, err := FirebaseMessageClient.UnsubscribeFromTopic(context.Background(), []string{userToken}, topic)
	if err != nil {
		log.Fatalf("error unsubscribing user from topic: %v\n", err)
		return err
	}
	return nil
}

func SendMessageToSubscriber(topic string, title string, body string, data map[string]string) error {

	message := &messaging.Message{
		// Title Body // R001 = ประกาศผล
		Data:  data,
		Topic: topic, // all_users
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		log.Fatalf("error sending message: %v\n", err)
		return err
	}

	return nil
}
