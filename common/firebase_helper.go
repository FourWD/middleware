package common

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
		LogError("FIREBASE_APP_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return nil, err
	}

	return app, nil
}
