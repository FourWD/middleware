package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken     = errors.New("invalid jwt token")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrRevokedToken     = errors.New("refresh token has been revoked")
)

// TokenUser carries the minimal identity needed for JWT generation.
// Convert your domain User to this before calling GeneratePair.
type TokenUser struct {
	ID    string
	Email string
	Role  string
}

// Claims is the JWT claims payload.
type Claims struct {
	UserID    string            `json:"user_id"`
	Email     string            `json:"email"`
	Role      string            `json:"role"`
	TokenType string            `json:"token_type"`
	Remark    map[string]string `json:"remark"`
	jwt.RegisteredClaims
}

// TokenPair holds an access/refresh token pair.
type TokenPair struct {
	AccessToken         string
	RefreshToken        string
	AccessExpiresInSec  int
	RefreshExpiresInSec int
	TokenType           string
}

// RefreshTokenStore persists refresh token state.
type RefreshTokenStore interface {
	Save(ctx context.Context, tokenID, username string, expiresAt time.Time) error
	IsActive(ctx context.Context, tokenID string) (bool, error)
	Revoke(ctx context.Context, tokenID string) error
}

// TokenManager issues and validates JWT tokens.
type TokenManager struct {
	cfg   AuthConfig
	store RefreshTokenStore
}

// NewTokenManager creates a TokenManager with the given auth config and store.
func NewTokenManager(cfg AuthConfig, store RefreshTokenStore) *TokenManager {
	return &TokenManager{
		cfg:   cfg,
		store: store,
	}
}

// GeneratePair creates an access + refresh token pair for the given user.
func (tm *TokenManager) GeneratePair(ctx context.Context, user *TokenUser) (*TokenPair, error) {
	accessToken, _, _, err := tm.generateToken(user, TokenTypeAccess, time.Duration(tm.cfg.AccessTokenTTLMinutes)*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshJTI, refreshExpiry, err := tm.generateToken(user, TokenTypeRefresh, time.Duration(tm.cfg.RefreshTokenTTLMinutes)*time.Minute)
	if err != nil {
		return nil, err
	}

	if err := tm.store.Save(ctx, refreshJTI, user.Email, refreshExpiry); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		AccessExpiresInSec:  tm.cfg.AccessTokenTTLMinutes * 60,
		RefreshExpiresInSec: tm.cfg.RefreshTokenTTLMinutes * 60,
		TokenType:           "Bearer",
	}, nil
}

func (tm *TokenManager) generateToken(user *TokenUser, tokenType string, ttl time.Duration) (string, string, time.Time, error) {
	now := time.Now()
	jti := uuid.NewString()
	expiry := now.Add(ttl)
	claims := Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: tokenType,
		Remark:    make(map[string]string),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    tm.cfg.JWTIssuer,
			Subject:   user.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(tm.cfg.JWTSecret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return signed, jti, expiry, nil
}

// Parse validates a JWT string and returns the claims.
func (tm *TokenManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected jwt signing method")
		}
		return []byte(tm.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ParseAccessToken validates a JWT string and ensures it is an access token.
func (tm *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	claims, err := tm.Parse(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// Refresh validates a refresh token, revokes it, and issues a new pair.
func (tm *TokenManager) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := tm.Parse(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	active, err := tm.store.IsActive(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrRevokedToken
	}

	if err := tm.store.Revoke(ctx, claims.ID); err != nil {
		return nil, err
	}

	return tm.GeneratePair(ctx, &TokenUser{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	})
}

// RevokeRefreshToken revokes a refresh token.
func (tm *TokenManager) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	claims, err := tm.Parse(refreshToken)
	if err != nil {
		return err
	}

	if claims.TokenType != TokenTypeRefresh {
		return ErrInvalidTokenType
	}

	return tm.store.Revoke(ctx, claims.ID)
}
