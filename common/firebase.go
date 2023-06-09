package common

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirebaseCtx context.Context
var FirebaseClient *firestore.Client

func ConnectFirebase(key string) {

	// Use a service account
	FirebaseCtx = context.Background()
	sa := option.WithCredentialsFile(key)
	app, err := firebase.NewApp(FirebaseCtx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	FirebaseClient, err = app.Firestore(FirebaseCtx)
	if err != nil {
		log.Fatalln(err)
	}

	// defer client.Close()

}
