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
	log.Println("ConnectMongo")
	clientOptions := options.Client().ApplyURI(key)
	log.Println("a")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("b")

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

	if client == nil {
		log.Fatal("Failed to create ctx options")
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Mongo ping: ", err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Mongo disconnect: ", err)
			log.Fatal(err)
		}
	}()

	log.Println("c")
	DatabaseMongo = new(MongoDB)
	DatabaseMongo.Client = client
	DatabaseMongo.Database = client.Database(databaseName)
	DatabaseMongo.Ctx = ctx

	log.Printf("CONNECT MONGO-DB SUCCESS [%s]", databaseName)
}
