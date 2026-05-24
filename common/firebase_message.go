package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/orm"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Timeouts are split so a slow Firebase call cannot eat the DB's budget.
const (
	firebaseCallTimeout = 30 * time.Second
	dbCallTimeout       = 10 * time.Second
)

func AddUserToSubscription(topic string, userID string, userToken string) error {
	if topic == "" || userID == "" || userToken == "" {
		return errors.New("topic, user ID, and user token are required")
	}

	requestID := uuid.NewString()
	start := time.Now()
	logData := map[string]any{
		"topic":   topic,
		"user_id": userID,
	}
	infra.AppLog.Event("FIREBASE_SUBSCRIBE_START", logData, requestID,
		infra.WithComponent(infra.ComponentFirebase),
		infra.WithOperation("subscribe"),
		infra.WithLogKind(infra.LogKindBusiness))

	fbCtx, fbCancel := context.WithTimeout(context.Background(), firebaseCallTimeout)
	defer fbCancel()

	if _, err := infra.FirebaseMessageClient.SubscribeToTopic(fbCtx, []string{userToken}, topic); err != nil {
		infra.AppLog.EventError(err, "FIREBASE_SUBSCRIBE_FAILURE", logData, requestID,
			infra.WithComponent(infra.ComponentFirebase),
			infra.WithOperation("subscribe"),
			infra.WithLogKind(infra.LogKindError))
		return fmt.Errorf("firebase subscribe: %w", err)
	}

	// Single transaction: upsert the topic row, then upsert the join.
	// FirstOrCreate keeps both operations idempotent — calling Add twice for
	// the same (topic, userID) is safe and creates no duplicates.
	//
	// NOTE: Firebase subscribe and the DB tx are NOT atomic. If the tx fails
	// after Firebase succeeds, the user is subscribed at Firebase but absent
	// from our DB. Run a reconciliation job out-of-band, or accept that a
	// subsequent Add call will heal the DB state (Firebase Subscribe is a
	// no-op when already subscribed).
	dbCtx, dbCancel := context.WithTimeout(context.Background(), dbCallTimeout)
	defer dbCancel()

	err := Database.WithContext(dbCtx).Transaction(func(tx *gorm.DB) error {
		var topicRow orm.NotificationTopic
		if err := tx.Where(orm.NotificationTopic{Name: topic}).
			Attrs(orm.NotificationTopic{ID: uuid.NewString()}).
			FirstOrCreate(&topicRow).Error; err != nil {
			return fmt.Errorf("upsert notification topic: %w", err)
		}

		var join orm.NotificationTopicUser
		if err := tx.Where(orm.NotificationTopicUser{
			NotificationTopicID: topicRow.ID,
			UserID:              userID,
		}).Attrs(orm.NotificationTopicUser{
			ID: uuid.NewString(),
		}).FirstOrCreate(&join).Error; err != nil {
			return fmt.Errorf("upsert notification topic user: %w", err)
		}
		return nil
	})
	if err != nil {
		infra.AppLog.EventError(err, "FIREBASE_SUBSCRIBE_PERSIST_FAILURE", logData, requestID,
			infra.WithComponent(infra.ComponentDB),
			infra.WithOperation("upsert_subscription"),
			infra.WithLogKind(infra.LogKindError))
		return err
	}

	logData["duration_ms"] = time.Since(start).Milliseconds()
	infra.AppLog.Event("FIREBASE_SUBSCRIBE_COMPLETE", logData, requestID,
		infra.WithComponent(infra.ComponentFirebase),
		infra.WithOperation("subscribe"),
		infra.WithLogKind(infra.LogKindBusiness))
	return nil
}

func RemoveUserFromSubscription(topic string, userID string, userToken string) error {
	if topic == "" || userID == "" || userToken == "" {
		return errors.New("topic, user ID, and user token are required")
	}

	requestID := uuid.NewString()
	start := time.Now()
	logData := map[string]any{
		"topic":   topic,
		"user_id": userID,
	}
	infra.AppLog.Event("FIREBASE_UNSUBSCRIBE_START", logData, requestID,
		infra.WithComponent(infra.ComponentFirebase),
		infra.WithOperation("unsubscribe"),
		infra.WithLogKind(infra.LogKindBusiness))

	// Firebase first, then DB. Order is intentional: if Firebase fails we
	// keep the DB row so a retry can re-attempt. If DB ops fail after
	// Firebase succeeded, a retry replays Firebase as a no-op and finishes
	// the DB cleanup — the flow is self-healing on retry.
	fbCtx, fbCancel := context.WithTimeout(context.Background(), firebaseCallTimeout)
	defer fbCancel()

	if _, err := infra.FirebaseMessageClient.UnsubscribeFromTopic(fbCtx, []string{userToken}, topic); err != nil {
		infra.AppLog.EventError(err, "FIREBASE_UNSUBSCRIBE_FAILURE", logData, requestID,
			infra.WithComponent(infra.ComponentFirebase),
			infra.WithOperation("unsubscribe"),
			infra.WithLogKind(infra.LogKindError))
		return fmt.Errorf("firebase unsubscribe: %w", err)
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), dbCallTimeout)
	defer dbCancel()
	db := Database.WithContext(dbCtx)

	var topicRow orm.NotificationTopic
	if err := db.Where("name = ?", topic).First(&topicRow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			infra.AppLog.EventWarn("FIREBASE_TOPIC_NOT_FOUND", logData, requestID,
				infra.WithComponent(infra.ComponentFirebase),
				infra.WithOperation("unsubscribe"),
				infra.WithLogKind(infra.LogKindDiagnostic))
			return fmt.Errorf("notification topic not found: %s", topic)
		}
		infra.AppLog.EventError(err, "FIREBASE_TOPIC_LOOKUP_FAILURE", logData, requestID,
			infra.WithComponent(infra.ComponentDB),
			infra.WithOperation("lookup"),
			infra.WithLogKind(infra.LogKindError),
			infra.WithField("table", "notification_topics"))
		return fmt.Errorf("lookup notification topic: %w", err)
	}
	logData["topic_id"] = topicRow.ID

	if err := db.Where("notification_topic_id = ? AND user_id = ?", topicRow.ID, userID).
		Unscoped().Delete(&orm.NotificationTopicUser{}).Error; err != nil {
		infra.AppLog.EventError(err, "FIREBASE_TOPIC_USER_DELETE_FAILURE", logData, requestID,
			infra.WithComponent(infra.ComponentDB),
			infra.WithOperation("delete"),
			infra.WithLogKind(infra.LogKindError),
			infra.WithField("table", "notification_topic_users"))
		return fmt.Errorf("delete notification topic user: %w", err)
	}

	logData["duration_ms"] = time.Since(start).Milliseconds()
	infra.AppLog.Event("FIREBASE_UNSUBSCRIBE_COMPLETE", logData, requestID,
		infra.WithComponent(infra.ComponentFirebase),
		infra.WithOperation("unsubscribe"),
		infra.WithLogKind(infra.LogKindBusiness))
	return nil
}
