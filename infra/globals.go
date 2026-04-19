package infra

import (
	"context"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"github.com/FourWD/middleware/model"
)

// GAEService returns the App Engine service name (set automatically as
// GAE_SERVICE by Google App Engine). Empty when not running on GAE.
func GAEService() string {
	return strings.TrimSpace(os.Getenv("GAE_SERVICE"))
}

// IsGAE reports whether the process is running on Google App Engine.
func IsGAE() bool {
	return GAEService() != ""
}

// Package-level state populated by NewApp so legacy code has a single
// well-known place to read app identity, logger, and third-party clients.
//
// Prefer passing dependencies via AppDeps for new code; these globals exist
// mainly to support code migrated from the old common.* pattern.
//
// AppInfo is named separately from the App type (the running application)
// to avoid a naming collision.
var (
	AppInfo model.AppInfo
	AppLog  *Logger

	FirebaseCtx           context.Context
	FirestoreClient       *firestore.Client
	FirebaseAuthClient    *firebaseAuth.Client
	FirebaseMessageClient *messaging.Client

	Mongo *MongoClient
)

func initGlobals(cfg CommonConfig, logger *Logger) {
	AppLog = logger
	AppInfo = model.AppInfo{
		Name:       cfg.AppID,
		Version:    cfg.AppVersion,
		Env:        cfg.AppEnv,
		GaeVersion: os.Getenv("GAE_VERSION"),
	}
}

func bindFirebaseGlobals(fc *FirebaseClient) {
	if fc == nil {
		return
	}
	FirebaseCtx = context.Background()
	if fc.Firestore != nil {
		FirestoreClient = fc.Firestore
	}
	if fc.Auth != nil {
		FirebaseAuthClient = fc.Auth
	}
	if fc.Notification != nil {
		FirebaseMessageClient = fc.Notification.Raw()
	}
}

func bindMongoGlobal(mc *MongoClient) {
	if mc == nil {
		return
	}
	Mongo = mc
}
