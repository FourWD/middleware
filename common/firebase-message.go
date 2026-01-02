package common

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/messaging"
	"github.com/FourWD/middleware/orm"

	"github.com/google/uuid"
)

var FirebaseMessageClient *messaging.Client

func ConnectFirebaseNotification(key string) error {
	app, err := initFirebaseApp(key)
	if err != nil {
		return err
	}

	FirebaseMessageClient, err = app.Messaging(context.Background())
	if err != nil {
		LogError("FIREBASE_MESSAGING_CLIENT_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return err
	}

	return nil
}

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
	requestID := uuid.NewString()
	logData := map[string]interface{}{
		"topic":     topic,
		"userID":    userID,
		"userToken": userToken,
	}
	Log("AddUserToSubscription", logData, requestID)

	if _, err := FirebaseMessageClient.SubscribeToTopic(context.Background(), []string{userToken}, topic); err != nil {
		LogError("FIREBASE_SUBSCRIBE_ERROR", logData, requestID)
		return err
	}

	newNotificationTopic := orm.NotificationTopic{
		ID:   uuid.NewString(),
		Name: topic,
	}
	if err := Database.Create(&newNotificationTopic).Error; err != nil {
		LogError("FIREBASE_INSERT_TOPIC_ERROR", logData, requestID)
	}

	var notificationTopic orm.NotificationTopic
	if err := Database.Where("name = ?", topic).First(&notificationTopic).Error; err != nil {
		LogError("FIREBASE_SELECT_TOPIC_ERROR", logData, requestID)
		return err
	}

	if notificationTopic.ID != "" {
		notificationTopicUser := orm.NotificationTopicUser{
			ID:                  uuid.NewString(),
			NotificationTopicID: notificationTopic.ID,
			UserID:              userID,
		}
		if err := Database.Create(&notificationTopicUser).Error; err != nil {
			LogError("FIREBASE_INSERT_TOPIC_USER_ERROR", logData, requestID)
			return err
		}
	}

	Log("AddUserToSubscription OK", logData, requestID)
	return nil
}

func RemoveUserFromSubscription(topic string, userID string, userToken string) error {
	logData := map[string]interface{}{
		"topic":     topic,
		"userID":    userID,
		"userToken": userToken,
	}
	Log("FIREBASE_REMOVE_USER_FROM_SUBSCRIPTION", logData, "")

	_, err := FirebaseMessageClient.UnsubscribeFromTopic(context.Background(), []string{userToken}, topic)
	if err != nil {
		LogError("FIREBASE_UNSUBSCRIBE_ERROR", map[string]interface{}{"error": err.Error(), "topic": topic, "userID": userID}, "")
		return err
	}
	var notificationTopicID string
	err = Database.Table("notification_topics").Select("id").Where("name = ?", topic).Scan(&notificationTopicID).Error
	if err != nil {
		LogError("FIREBASE_TOPIC_ID_ERROR", map[string]interface{}{"error": err.Error(), "topic": topic}, "")
		return err
	}

	if notificationTopicID == "" {
		LogWarning("FIREBASE_TOPIC_ID_NOT_FOUND", map[string]interface{}{"topic": topic}, "")
		return fmt.Errorf("notification topic ID not found for topic: %s", topic)
	}

	err = Database.Where("notification_topic_id = ? AND user_id = ?", notificationTopicID, userID).Unscoped().Delete(&orm.NotificationTopicUser{}).Error
	if err != nil {
		LogError("FIREBASE_REMOVE_USER_DB_ERROR", map[string]interface{}{"error": err.Error(), "topic": topic, "userID": userID}, "")
		return err
	}

	return nil
}

func SendMessageToSubscriber(topic string, title string, body string, data map[string]string) error {
	logData := map[string]interface{}{
		"topic": topic,
		"title": title,
		"body":  body,
	}
	Log("FIREBASE_SEND_MESSAGE_TO_SUBSCRIBER", logData, "")

	message := &messaging.Message{
		Data:  data,
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err := FirebaseMessageClient.Send(context.Background(), message)
	if err != nil {
		LogError("FIREBASE_SEND_MESSAGE_ERROR", map[string]interface{}{"error": err.Error(), "topic": topic}, "")
		return err
	}

	return nil
}
