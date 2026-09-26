package infra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const legacyBlacklistMinTTL = 3 * 24 * time.Hour

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
// blacklist_tokens collection until max(now+3d, token exp). The caller supplies ctx
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
		"expiresAt":  legacyBlacklistExpiry(jwtToken, now),
	})
	return err
}

// legacyBlacklistExpiry keeps the entry at least until the token's own exp
// so a long-lived token cannot outlive its blacklist row. The exp is read
// unverified; unparsable input falls back to the 3-day minimum.
func legacyBlacklistExpiry(jwtToken string, now time.Time) time.Time {
	expiresAt := now.Add(legacyBlacklistMinTTL)
	raw := strings.TrimSpace(jwtToken)
	if t, err := ParseBearerToken(raw); err == nil {
		raw = strings.TrimSpace(t)
	}
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(raw, claims); err != nil {
		return expiresAt
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return expiresAt
	}
	if exp.After(expiresAt) {
		return exp.Time
	}
	return expiresAt
}
