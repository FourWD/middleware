package common

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func BlacklistJwtToken(jwtToken string) error {
	if jwtToken == "" {
		return errors.New("no token")
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(3 * 24 * time.Hour)

	collection := DatabaseMongoMiddleware.Database.Collection("blacklist_tokens")
	data := bson.M{
		"token":     jwtToken,
		"createdAt": createdAt,
		"expiresAt": expiresAt,
	}

	_, err := collection.InsertOne(context.TODO(), data)
	return err
}
