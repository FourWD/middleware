package common

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DatabaseMongo *MongoDB
var DatabaseMongoMiddleware *MongoDB

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	Ctx      context.Context
}

func connectMongoInternal(key string, databaseName string) *MongoDB {
	clientOptions := options.Client().ApplyURI(key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if clientOptions == nil {
		LogError("MONGO_CLIENT_OPTIONS_ERROR", map[string]interface{}{"database": databaseName}, "")
		panic("Failed to create client options")
	}

	if ctx == nil {
		LogError("MONGO_CTX_ERROR", map[string]interface{}{"database": databaseName}, "")
		panic("Failed to create ctx options")
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		LogError("MONGO_CONNECT_ERROR", map[string]interface{}{"error": err.Error(), "database": databaseName}, "")
		panic(err)
	}

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
