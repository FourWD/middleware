package infra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// hashBlacklistToken returns the hex SHA-256 of the JWT. The blacklist
// stores the hash, not the token itself, so an accidental DB/Redis dump
// does not expose live bearer tokens. Lookups re-compute the hash before
// querying — the function is deterministic.
//
// Used by every blacklist read/write site: BlacklistJwtToken (legacy
// free function), MongoBlacklistStore, RedisBlacklistStore, and the
// IsJwtValid lookup in auth_middleware.go. Keep them aligned — a
// mismatched hash on read vs write silently breaks revocation.
func hashBlacklistToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// BlacklistJwtToken stores the SHA-256 hash of jwtToken in the
// blacklist_tokens collection with a 3-day TTL. The caller supplies ctx
// so the write inherits the request's deadline and trace.
//
// Migration note: pre-existing documents using the plaintext "token"
// field are orphaned — they'll TTL-expire within 3 days. The legacy
// "token_unique" Mongo index can be dropped after that window.
func BlacklistJwtToken(ctx context.Context, jwtToken string) error {
	if jwtToken == "" {
		return errors.New("no token")
	}
	if MongoMiddleware == nil {
		return errors.New("mongo middleware client not initialized")
	}

	now := time.Now()
	_, err := MongoMiddleware.Database().Collection("blacklist_tokens").InsertOne(ctx, bson.M{
		"token_hash": hashBlacklistToken(jwtToken),
		"createdAt":  now,
		"expiresAt":  now.Add(3 * 24 * time.Hour),
	})
	return err
}
