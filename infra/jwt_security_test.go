package infra

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const b64URLAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// malleateSignature swaps the last signature char for one that differs only
// in the unused low bits, so lax decoders yield identical signature bytes.
func malleateSignature(t *testing.T, token string) string {
	t.Helper()
	idx := strings.IndexByte(b64URLAlphabet, token[len(token)-1])
	if idx < 0 {
		t.Fatalf("unexpected signature char %q", token[len(token)-1])
	}
	alt := token[:len(token)-1] + string(b64URLAlphabet[idx^1])

	sig := func(s string) []byte {
		b, err := base64.RawURLEncoding.DecodeString(s[strings.LastIndexByte(s, '.')+1:])
		if err != nil {
			t.Fatalf("lax decode: %v", err)
		}
		return b
	}
	if string(sig(token)) != string(sig(alt)) {
		t.Fatal("alternate char must decode to the same bytes under lax decoding")
	}
	return alt
}

func testTokenManager(store RefreshTokenStore) *TokenManager {
	return NewTokenManager(AuthConfig{
		JWTSecret: "test-secret", JWTIssuer: "test", AccessTokenTTLMinutes: 5, RefreshTokenTTLMinutes: 10,
	}, store)
}

func TestTokenManagerParse_RejectsMalleatedSignature(t *testing.T) {
	tm := testTokenManager(&refreshTokenStoreMock{})
	pair, err := tm.GeneratePair(context.Background(), &TokenUser{ID: "1", Email: "a@example.com", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tm.Parse(pair.AccessToken); err != nil {
		t.Fatalf("canonical token must parse: %v", err)
	}
	if _, err := tm.Parse(malleateSignature(t, pair.AccessToken)); err == nil {
		t.Fatal("malleated signature must be rejected")
	}
}

func TestLegacyJWT_RejectsMalleatedSignature(t *testing.T) {
	ensureJWTSecret(t)
	token, err := GenerateJWTToken("user-1", "admin", nil, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	alt := malleateSignature(t, token)

	if _, err := extractClaimFromToken(token, "user_id", false); err != nil {
		t.Fatalf("canonical token must parse: %v", err)
	}
	if _, err := extractClaimFromToken(alt, "user_id", false); err == nil {
		t.Fatal("extractClaimFromToken must reject malleated signature")
	}

	app := fiber.New()
	app.Get("/", checkAuth, func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	for tok, want := range map[string]int{token: http.StatusOK, alt: http.StatusUnauthorized} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != want {
			t.Fatalf("checkAuth status = %d, want %d", res.StatusCode, want)
		}
	}
}

// atomicRefreshStore mimics the built-in stores: Revoke reports
// ErrRevokedToken when nothing was deleted.
type atomicRefreshStore struct {
	mu     sync.Mutex
	active map[string]bool
	// gate, when set, holds every IsActive caller until all have passed the
	// check, reproducing the check-then-act window deterministically.
	gate *sync.WaitGroup
}

func (s *atomicRefreshStore) Save(_ context.Context, id, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		s.active = map[string]bool{}
	}
	s.active[id] = true
	return nil
}

func (s *atomicRefreshStore) IsActive(_ context.Context, id string) (bool, error) {
	s.mu.Lock()
	ok := s.active[id]
	gate := s.gate
	s.mu.Unlock()
	if gate != nil {
		gate.Done()
		gate.Wait()
	}
	return ok, nil
}

func (s *atomicRefreshStore) Revoke(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active[id] {
		return ErrRevokedToken
	}
	delete(s.active, id)
	return nil
}

func TestRefresh_ConcurrentReplayYieldsSinglePair(t *testing.T) {
	const n = 8
	store := &atomicRefreshStore{}
	tm := testTokenManager(store)
	pair, err := tm.GeneratePair(context.Background(), &TokenUser{ID: "1", Email: "a@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	store.gate = &sync.WaitGroup{}
	store.gate.Add(n)

	var ok, revoked atomic.Int32
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := tm.Refresh(context.Background(), pair.RefreshToken)
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, ErrRevokedToken):
				revoked.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 1 || revoked.Load() != n-1 {
		t.Fatalf("ok=%d revoked=%d, want 1 and %d", ok.Load(), revoked.Load(), n-1)
	}
}

func TestRefresh_SerialReplayFails(t *testing.T) {
	tm := testTokenManager(&atomicRefreshStore{})
	ctx := context.Background()
	pair, err := tm.GeneratePair(ctx, &TokenUser{ID: "1", Email: "a@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tm.Refresh(ctx, pair.RefreshToken); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if _, err := tm.Refresh(ctx, pair.RefreshToken); !errors.Is(err, ErrRevokedToken) {
		t.Fatalf("second refresh err = %v, want ErrRevokedToken", err)
	}
	if err := tm.RevokeRefreshToken(ctx, pair.RefreshToken); err != nil {
		t.Fatalf("logout of already-revoked token must stay idempotent: %v", err)
	}
}

func TestLegacyBlacklistExpiry(t *testing.T) {
	now := time.Now()
	minExp := now.Add(legacyBlacklistMinTTL)
	sign := func(claims jwt.MapClaims) string {
		s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("k"))
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	long := time.Unix(now.Add(30*24*time.Hour).Unix(), 0)

	cases := map[string]struct {
		in   string
		want time.Time
	}{
		"garbage":       {"not-a-jwt", minExp},
		"empty bearer":  {"Bearer ", minExp},
		"short exp":     {sign(jwt.MapClaims{"exp": now.Add(time.Hour).Unix()}), minExp},
		"long exp":      {sign(jwt.MapClaims{"exp": long.Unix()}), long},
		"bearer prefix": {"Bearer " + sign(jwt.MapClaims{"exp": long.Unix()}), long},
		"no exp":        {sign(jwt.MapClaims{"sub": "x"}), minExp},
	}
	for name, tc := range cases {
		if got := legacyBlacklistExpiry(tc.in, now); !got.Equal(tc.want) {
			t.Errorf("%s: got %v want %v", name, got, tc.want)
		}
	}
}

func TestGoogleJWKS_UnknownKidRateLimited(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	var kids atomic.Value
	kids.Store([]string{"k1"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		var keys []map[string]string
		for _, kid := range kids.Load().([]string) {
			keys = append(keys, map[string]string{
				"kid": kid, "kty": "RSA",
				"n": base64.RawURLEncoding.EncodeToString(priv.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(priv.E)).Bytes()),
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	}))
	defer srv.Close()

	const interval = 200 * time.Millisecond
	c := &googleJWKSCache{url: srv.URL, minInterval: interval}
	ctx := context.Background()

	spray := func() {
		var wg sync.WaitGroup
		for i := range 50 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := c.key(ctx, fmt.Sprintf("random-%d", i)); err == nil {
					t.Error("unknown kid must fail")
				}
			}()
		}
		wg.Wait()
	}

	if _, err := c.key(ctx, "k1"); err != nil {
		t.Fatal(err)
	}
	spray()
	if got := hits.Load(); got != 1 {
		t.Fatalf("unknown kids within interval must not fetch, fetches = %d", got)
	}

	time.Sleep(interval + 50*time.Millisecond)
	spray()
	if got := hits.Load(); got != 2 {
		t.Fatalf("burst after interval must fetch once, fetches = %d", got)
	}

	// Rotation: a genuinely new kid is picked up once the interval elapses.
	kids.Store([]string{"k1", "k2"})
	if _, err := c.key(ctx, "k2"); err == nil {
		t.Fatal("k2 should not be visible before the interval elapses")
	}
	time.Sleep(interval + 50*time.Millisecond)
	if _, err := c.key(ctx, "k2"); err != nil {
		t.Fatalf("rotated kid must be picked up: %v", err)
	}
	if _, err := c.key(ctx, "k1"); err != nil {
		t.Fatal(err)
	}
	if got := hits.Load(); got != 3 {
		t.Fatalf("fetches = %d, want 3", got)
	}
}
