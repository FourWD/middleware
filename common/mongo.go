package common

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DatabaseMongo *MongoDB

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	Ctx      context.Context
}

func ConnectMongo(key string, databaseName string) {
	clientOptions := options.Client().ApplyURI(key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	DatabaseMongo.Client = client
	DatabaseMongo.Database = client.Database(databaseName)
	DatabaseMongo.Ctx = ctx

	log.Printf("Connected to MongoDB! [%s]", databaseName)
}
