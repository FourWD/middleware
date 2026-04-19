package infra

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func initFirebaseApp(credentialsFile string) (*firebase.App, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		AppLog.EventError(err, "FIREBASE_APP_ERROR", nil, "")
		return nil, err
	}

	return app, nil
}
