package common

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoCtx context.Context
var MongoClient *mongo.Client

func ConnectMongo(key string) {
	clientOptions := options.Client().ApplyURI(key)
	MongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	MongoClient, err := mongo.Connect(MongoCtx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = MongoClient.Disconnect(MongoCtx); err != nil {
			log.Fatal(err)
		}
	}()
}
