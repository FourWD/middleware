package infra

import (
	"context"
	"errors"
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
	// FindOne with an _id-only projection short-circuits on first match;
	// CountDocuments would scan-and-count even after the first hit.
	opts := options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}})
	err := s.collection.FindOne(ctx, bson.M{"token_hash": hashBlacklistToken(token)}, opts).Err()
	if err == nil {
		return true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	return false, err
}

func (s *MongoBlacklistStore) Add(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := s.collection.InsertOne(ctx, bson.M{
		"token_hash": hashBlacklistToken(token),
		"createdAt":  time.Now(),
		"expiresAt":  expiresAt,
	})
	return err
}

// EnsureIndexes creates the unique lookup + TTL indexes the store needs.
// Idempotent — safe to call on every boot. Without the TTL index the
// collection grows without bound; without the unique index on
// `token_hash`, IsBlacklisted does a collection scan that gets linearly
// slower as the table fills up.
//
// Caller should treat errors as non-fatal: a Mongo user without
// createIndex permission, a pre-existing index with conflicting options,
// or a transient Mongo blip should not block the service from booting.
// Log the error and let the operator reconcile indexes manually.
//
// Migration note: a pre-existing `token_unique` index (from the
// plaintext-token era) still references the now-empty `token` field.
// Drop it manually after the 3-day TTL window has fully cleared the
// legacy rows.
func (s *MongoBlacklistStore) EnsureIndexes(ctx context.Context) error {
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("token_hash_unique"),
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
	n, err := s.client.Exists(ctx, redisBlacklistKey(token)).Result()
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
	return s.client.Set(ctx, redisBlacklistKey(token), 1, ttl).Err()
}

// redisBlacklistKey builds the Redis key for a blacklisted token. Uses
// the hashed token (not raw) so a Redis MONITOR / RDB dump cannot leak
// live bearer tokens.
func redisBlacklistKey(token string) string {
	return "blacklist:" + hashBlacklistToken(token)
}
