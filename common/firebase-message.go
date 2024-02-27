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

func SendMessageToUser(userToken string, data map[string]string) error { // 1 : 1
	// Title Body
	// message := &messaging.Message{
	// 	Data:  *data,
	// 	Token: *userToken,
	// }

	_, err := FirebaseMessageClient.Send(context.Background(), &messaging.Message{
		Data:  data,
		Token: userToken,
	})
	if err != nil {
		log.Fatalf("error sending message: %v\n", err)
		return err
	}

	return nil
}

func AddUserToSubscription(userToken string, topic string) error { // เอาคน (topic) เข้า กรุป auction
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

func SendMessageToSubscriber(topic string, data map[string]string) error {
	message := &messaging.Message{
		// Title Body // R001 = ประกาศผล
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
