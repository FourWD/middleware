package infra

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type refreshTokenStoreMock struct {
	active map[string]bool
}

func (m *refreshTokenStoreMock) Save(_ context.Context, tokenID, _ string, _ time.Time) error {
	if m.active == nil {
		m.active = make(map[string]bool)
	}

	m.active[tokenID] = true
	return nil
}

func (m *refreshTokenStoreMock) IsActive(_ context.Context, tokenID string) (bool, error) {
	return m.active[tokenID], nil
}

func (m *refreshTokenStoreMock) Revoke(_ context.Context, tokenID string) error {
	delete(m.active, tokenID)
	return nil
}

func TestTokenManager_GeneratePairAndRefresh(t *testing.T) {
	store := &refreshTokenStoreMock{}
	manager := NewTokenManager(AuthConfig{
		JWTSecret:              "test-secret",
		JWTIssuer:              "test-suite",
		AccessTokenTTLMinutes:  5,
		RefreshTokenTTLMinutes: 10,
	}, store)

	user := &TokenUser{
		ID:    "7",
		Email: "admin@example.com",
		Role:  "admin",
	}

	ctx := context.Background()

	pair, err := manager.GeneratePair(ctx, user)
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)

	accessClaims, err := manager.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, accessClaims.UserID)
	assert.Equal(t, user.Email, accessClaims.Email)
	assert.Equal(t, user.Role, accessClaims.Role)

	refreshed, err := manager.Refresh(ctx, pair.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, pair.RefreshToken, refreshed.RefreshToken)
	assert.NotEmpty(t, refreshed.AccessToken)

	_, err = manager.Refresh(ctx, pair.RefreshToken)
	require.ErrorIs(t, err, ErrRevokedToken)
}
