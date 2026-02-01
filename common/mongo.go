package common

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DatabaseMongo *MongoDB
var DatabaseMongoMiddleware *MongoDB

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	Ctx      context.Context
}

func connectMongoInternal(key string, databaseName string) *MongoDB {
	clientOptions := options.Client().ApplyURI(key).SetTimeout(10 * time.Second)

	if clientOptions == nil {
		LogError("MONGO_CLIENT_OPTIONS_ERROR", map[string]interface{}{"database": databaseName}, "")
		panic("Failed to create client options")
	}

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		LogError("MONGO_CONNECT_ERROR", map[string]interface{}{"error": err.Error(), "database": databaseName}, "")
		panic(err)
	}

	ctx := context.Background()
	if err = client.Ping(ctx, nil); err != nil {
		LogError("MONGO_PING_ERROR", map[string]interface{}{"error": err.Error(), "database": databaseName}, "")
		panic(err)
	}

	return &MongoDB{
		Client:   client,
		Database: client.Database(databaseName),
		Ctx:      ctx,
	}
}

func ConnectMongo(key string, databaseName string) {
	DatabaseMongo = connectMongoInternal(key, databaseName)
	Log("MONGO_CONNECTION_SUCCESS", map[string]interface{}{"database": databaseName}, "")
}

func ConnectMongoMiddleware(key string, databaseName string) {
	DatabaseMongoMiddleware = connectMongoInternal(key, databaseName)
	Log("MONGO_MIDDLEWARE_CONNECTION_SUCCESS", map[string]interface{}{"database": databaseName}, "")
}
