package infra

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
