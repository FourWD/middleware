package infra

import "testing"

// FIREBASE_NOTIFICATION_CREDENTIALS and PUBSUB_PROJECT_ID were removed: FCM
// shares the Firestore credentials and Pub/Sub uses the service's own project.
// Both tests set the retired variable to prove it no longer has any effect.

func TestLoadCommonConfig_FirebaseNotificationUsesCredentials(t *testing.T) {
	t.Setenv("FIREBASE_CREDENTIALS", "firebase-key.json")
	t.Setenv("FIREBASE_NOTIFICATION_CREDENTIALS", "retired.json")

	if got := LoadCommonConfig().Firebase.NotificationCredentialsFile; got != "firebase-key.json" {
		t.Fatalf("NotificationCredentialsFile = %q, want firebase-key.json", got)
	}
}

func TestLoadCommonConfig_PubSubUsesGoogleCloudProject(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "auction-dev-3a43a")
	t.Setenv("PUBSUB_PROJECT_ID", "retired-project")

	if got := LoadCommonConfig().PubSub.ProjectID; got != "auction-dev-3a43a" {
		t.Fatalf("ProjectID = %q, want auction-dev-3a43a", got)
	}
}
