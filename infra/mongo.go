package infra

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoConfig holds MongoDB connection configuration.
type MongoConfig struct {
	URI      string
	Database string
}

// LoadMongoConfig reads the primary MongoDB configuration from environment variables.
func LoadMongoConfig() MongoConfig {
	return MongoConfig{
		URI:      GetEnv("MONGO_URI", ""),
		Database: GetEnv("MONGO_DATABASE", "middleware"),
	}
}

// LoadMongoMiddlewareConfig reads the dedicated middleware MongoDB configuration.
// Use this when auth/blacklist data should live in a cluster separate from
// business data.
func LoadMongoMiddlewareConfig() MongoConfig {
	return MongoConfig{
		URI:      GetEnv("MONGO_MIDDLEWARE_URI", ""),
		Database: GetEnv("MONGO_MIDDLEWARE_DATABASE", "middleware"),
	}
}

// MongoClient wraps the official MongoDB client with a selected database.
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

// ConnectMongo creates a new MongoDB client and pings the server.
func ConnectMongo(ctx context.Context, cfg MongoConfig) (*MongoClient, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &MongoClient{
		client:   client,
		database: client.Database(cfg.Database),
	}, nil
}

// Collection returns a handle to the named collection.
func (mc *MongoClient) Collection(name string) *mongo.Collection {
	return mc.database.Collection(name)
}

// Client returns the underlying *mongo.Client.
func (mc *MongoClient) Client() *mongo.Client {
	if mc == nil {
		return nil
	}
	return mc.client
}

// Database returns the selected *mongo.Database.
func (mc *MongoClient) Database() *mongo.Database {
	if mc == nil {
		return nil
	}
	return mc.database
}

// Close disconnects the MongoDB client.
func (mc *MongoClient) Close(ctx context.Context) error {
	if mc.client != nil {
		return mc.client.Disconnect(ctx)
	}
	return nil
}
