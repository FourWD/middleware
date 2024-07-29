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

	if clientOptions == nil {
		log.Fatal("Failed to create client options")
	}

	if ctx == nil {
		log.Fatal("Failed to create ctx options")
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Mongo connect: ", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Mongo ping: ", err)
	}

	DatabaseMongo = new(MongoDB)
	DatabaseMongo.Client = client
	DatabaseMongo.Database = client.Database(databaseName)
	DatabaseMongo.Ctx = ctx

	log.Printf("CONNECT MONGO-DB SUCCESS [%s]", databaseName)
}
