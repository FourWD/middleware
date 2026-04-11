package infra

import (
	"context"
	"fmt"
	"time"
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
