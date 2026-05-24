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
		AppLog.EventError(err, "FIREBASE_APP_INIT_FAILURE", nil, "",
			WithComponent(ComponentFirebase),
			WithOperation("init_app"),
			WithLogKind(LogKindError))
		return nil, err
	}

	return app, nil
}
