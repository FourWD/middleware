package common

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/FourWD/middleware/orm"

	"github.com/google/uuid"
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
		log.Printf("error getting Messaging client: %v\n", err)
		return err
	}

	return nil
}

// config struct message
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

func SendMessageToUser(userToken string, title string, body string, data map[string]string) error {
	// Access title and body directly from the data map
	//
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

	result, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		return fmt.Errorf("error sending message: %s, %v", result, err)
	}

	return nil
}

func AddUserToSubscription(topic string, userID string, userToken string) error {
	if _, err := FirebaseMessageClient.SubscribeToTopic(context.Background(), []string{userToken}, topic); err != nil {
		log.Println("error subscribing user to topic: ", err)
	}
	// ========================================================================================
	newNotificationTopic := orm.NotificationTopic{
		ID:   uuid.NewString(),
		Name: topic,
	}
	if err := Database.Create(&newNotificationTopic).Error; err != nil {
		log.Println("failed to insert notification topic: ", err)
	}
	// ========================================================================================
	var notificationTopic orm.NotificationTopic
	if err := Database.Where("name = ?", topic).First(&notificationTopic).Error; err != nil {
		log.Println("failed to select notification topic: ", err)
	}
	// ========================================================================================
	if notificationTopic.ID != "" {
		notificationTopicUser := orm.NotificationTopicUser{
			ID:                  uuid.NewString(),
			NotificationTopicID: notificationTopic.ID,
			UserID:              userID,
		}
		if err := Database.Create(&notificationTopicUser).Error; err != nil {
			log.Println("failed to insert notification user topic user: ", err)
		}
	}
	// ========================================================================================
	return nil
}

func RemoveUserFromSubscription(topic string, userID string, userToken string) error {
	fmt.Printf("UserToken: %s, userID: %s, Topic: %s\n", userToken, userID, topic)

	_, err := FirebaseMessageClient.UnsubscribeFromTopic(context.Background(), []string{userToken}, topic)
	if err != nil {
		log.Printf("error unsubscribing user from topic: %v\n", err)
		return err
	}
	var notificationTopicID string
	err = Database.Table("notification_topics").Select("id").Where("name = ?", topic).Scan(&notificationTopicID).Error
	if err != nil {
		log.Printf("error finding notification topic ID: %v\n", err)
		return err
	}

	if notificationTopicID == "" {
		log.Printf("notification topic ID not found for topic: %s\n", topic)
		return fmt.Errorf("notification topic ID not found for topic: %s", topic)
	}

	err = Database.Where("notification_topic_id = ? AND user_id = ?", notificationTopicID, userID).Unscoped().Debug().Delete(&orm.NotificationTopicUser{}).Error
	if err != nil {
		log.Printf("error removing user from topic in database: %v\n", err)
		return err
	}

	return nil
}

func SendMessageToSubscriber(topic string, title string, body string, data map[string]string) error {
	Print("data", title)
	message := &messaging.Message{
		Data:  data,
		Topic: topic, // all_users
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("error sending message: %v\n", err)
		return err
	}

	return nil
}
