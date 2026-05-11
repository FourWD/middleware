package infra

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BlacklistStore checks and manages token blacklisting.
type BlacklistStore interface {
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	Add(ctx context.Context, token string, expiresAt time.Time) error
}

// MongoBlacklistStore implements BlacklistStore using MongoDB.
// Collection: blacklist_tokens — matching the old middleware behavior.
type MongoBlacklistStore struct {
	collection *mongo.Collection
}

// NewMongoBlacklistStore creates a store backed by the blacklist_tokens collection.
func NewMongoBlacklistStore(client *MongoClient) *MongoBlacklistStore {
	return &MongoBlacklistStore{
		collection: client.Collection("blacklist_tokens"),
	}
}

func (s *MongoBlacklistStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	count, err := s.collection.CountDocuments(ctx, bson.M{"token": token})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *MongoBlacklistStore) Add(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := s.collection.InsertOne(ctx, bson.M{
		"token":     token,
		"createdAt": time.Now(),
		"expiresAt": expiresAt,
	})
	return err
}

// EnsureIndexes creates the unique lookup + TTL indexes the store needs.
// Idempotent — safe to call on every boot. Without the TTL index the
// collection grows without bound; without the unique index on `token`,
// IsBlacklisted does a collection scan that gets linearly slower as the
// table fills up.
//
// Caller should treat errors as non-fatal: a Mongo user without
// createIndex permission, a pre-existing index with conflicting options,
// or a transient Mongo blip should not block the service from booting.
// Log the error and let the operator reconcile indexes manually.
func (s *MongoBlacklistStore) EnsureIndexes(ctx context.Context) error {
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("token_unique"),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("expiresAt_ttl"),
		},
	})
	return err
}

// RedisBlacklistStore implements BlacklistStore using Redis.
// Keys are stored as "blacklist:<token>" with TTL matching the token expiry.
type RedisBlacklistStore struct {
	client *RedisClient
}

// NewRedisBlacklistStore creates a blacklist store backed by Redis.
func NewRedisBlacklistStore(client *RedisClient) *RedisBlacklistStore {
	return &RedisBlacklistStore{client: client}
}

func (s *RedisBlacklistStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	n, err := s.client.Exists(ctx, fmt.Sprintf("blacklist:%s", token)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *RedisBlacklistStore) Add(ctx context.Context, token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, fmt.Sprintf("blacklist:%s", token), 1, ttl).Err()
}
