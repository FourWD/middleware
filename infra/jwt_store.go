package infra

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RedisRefreshTokenStore implements RefreshTokenStore using Redis.
type RedisRefreshTokenStore struct {
	client *RedisClient
	prefix string
}

// NewRedisRefreshTokenStore creates a store keyed by JWT issuer.
func NewRedisRefreshTokenStore(client *RedisClient, cfg AuthConfig) *RedisRefreshTokenStore {
	return &RedisRefreshTokenStore{
		client: client,
		prefix: fmt.Sprintf("%s:refresh_token:", cfg.JWTIssuer),
	}
}

func (s *RedisRefreshTokenStore) Save(ctx context.Context, tokenID, username string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return ErrInvalidToken
	}

	return s.client.Set(ctx, s.key(tokenID), username, ttl).Err()
}

func (s *RedisRefreshTokenStore) IsActive(ctx context.Context, tokenID string) (bool, error) {
	count, err := s.client.Exists(ctx, s.key(tokenID)).Result()
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

func (s *RedisRefreshTokenStore) Revoke(ctx context.Context, tokenID string) error {
	return s.client.Del(ctx, s.key(tokenID)).Err()
}

func (s *RedisRefreshTokenStore) key(tokenID string) string {
	return s.prefix + tokenID
}

// MongoRefreshTokenStore implements RefreshTokenStore using MongoDB.
// Collection: refresh_tokens — sits next to the blacklist_tokens collection
// in the middleware Mongo database so operators only manage one DB for
// auth state. Mirrors RedisRefreshTokenStore's API; pick whichever backend
// matches what's already enabled in the project (MONGO_MIDDLEWARE_ENABLED
// vs REDIS_ENABLED).
type MongoRefreshTokenStore struct {
	collection *mongo.Collection
	issuer     string
}

// NewMongoRefreshTokenStore creates a store keyed by JWT issuer. The
// issuer is stamped into every doc so multiple services sharing one Mongo
// cluster don't accidentally validate each other's refresh tokens.
func NewMongoRefreshTokenStore(client *MongoClient, cfg AuthConfig) *MongoRefreshTokenStore {
	return &MongoRefreshTokenStore{
		collection: client.Collection("refresh_tokens"),
		issuer:     cfg.JWTIssuer,
	}
}

func (s *MongoRefreshTokenStore) Save(ctx context.Context, tokenID, username string, expiresAt time.Time) error {
	if time.Until(expiresAt) <= 0 {
		return ErrInvalidToken
	}
	_, err := s.collection.InsertOne(ctx, bson.M{
		"token_id":   tokenID,
		"issuer":     s.issuer,
		"user_email": username,
		"expires_at": expiresAt,
		"created_at": time.Now().UTC(),
	})
	return err
}

func (s *MongoRefreshTokenStore) IsActive(ctx context.Context, tokenID string) (bool, error) {
	// Filter on issuer so cross-service token collisions are impossible
	// even if a future migration loosens the unique index. The TTL index
	// purges expired rows asynchronously; the explicit expires_at gate
	// below covers the window between expiry and the sweep.
	n, err := s.collection.CountDocuments(ctx, bson.M{
		"token_id":   tokenID,
		"issuer":     s.issuer,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *MongoRefreshTokenStore) Revoke(ctx context.Context, tokenID string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{
		"token_id": tokenID,
		"issuer":   s.issuer,
	})
	return err
}

// EnsureIndexes creates the unique lookup + TTL indexes the store needs.
// Idempotent — safe to call on every boot. See MongoBlacklistStore's
// EnsureIndexes for the same fail-soft contract: callers should log and
// continue on error, never abort startup.
func (s *MongoRefreshTokenStore) EnsureIndexes(ctx context.Context) error {
	_, err := s.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "token_id", Value: 1},
				{Key: "issuer", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("token_issuer_unique"),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("expires_at_ttl"),
		},
	})
	return err
}
