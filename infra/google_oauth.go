// Package-level: Google OIDC ID-token verifier.
//
// Verifies tokens minted by Google Identity Services (the credential
// returned by `google.accounts.id.renderButton(...)` callback). Uses
// only the standard library — no `google.golang.org/api/idtoken` or
// Firebase Admin SDK pull-in, so the binary footprint stays tiny.
//
// Typical wiring:
//
//	verifier := infra.NewGoogleVerifier(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
//	claims, err := verifier.Verify(ctx, body.IDToken)
//	if err != nil { return WriteError(c, 401, ErrBadRequestWrap("invalid_token", err)) }
//	// claims.Email is the verified user identity (Gmail or Workspace).

package infra

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GoogleClaims is the subset of an ID-token payload we expose to callers.
// EmailVerified must be true for Verify to succeed — anything else is
// rejected, so a drive-by Gmail account can't sign in as someone else's
// address by faking the email claim.
type GoogleClaims struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// GoogleVerifier validates Google-issued OIDC ID tokens against the
// public JWKs at https://www.googleapis.com/oauth2/v3/certs.
//
// Verify checks:
//   - 3-segment JWT shape, RS256 alg, kid present in JWKs
//   - signature with the matching JWKs RSA public key
//   - iss in {accounts.google.com, https://accounts.google.com}
//   - aud == configured clientID (single string OR array containing it)
//   - exp in the future
//   - email_verified == true
//   - sub + email non-empty
//
// JWKs are cached for one hour; cache misses (or stale entries) trigger
// a synchronous refresh. Google rotates keys roughly every 6h, so 1h is
// a comfortably-tight TTL.
type GoogleVerifier struct {
	clientID string
	jwks     *googleJWKSCache
}

// NewGoogleVerifier returns a verifier for the given OAuth client id.
// The id is the "Client ID for Web application" in Google Cloud
// Console → APIs & Services → Credentials.
func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		clientID: clientID,
		jwks: &googleJWKSCache{
			url: "https://www.googleapis.com/oauth2/v3/certs",
		},
	}
}

func (v *GoogleVerifier) Verify(ctx context.Context, idToken string) (*GoogleClaims, error) {
	if v.clientID == "" {
		return nil, errors.New("google verifier: client id not configured")
	}
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("google id token: not a JWT")
	}
	hb, err := googleB64Decode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("google id token: header decode: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(hb, &header); err != nil {
		return nil, fmt.Errorf("google id token: header parse: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("google id token: unsupported alg %q", header.Alg)
	}

	pub, err := v.jwks.key(ctx, header.Kid)
	if err != nil {
		return nil, fmt.Errorf("google id token: jwks: %w", err)
	}

	signingInput := []byte(parts[0] + "." + parts[1])
	sig, err := googleB64Decode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("google id token: signature decode: %w", err)
	}
	hashed := sha256.Sum256(signingInput)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig); err != nil {
		return nil, fmt.Errorf("google id token: signature: %w", err)
	}

	pb, err := googleB64Decode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("google id token: payload decode: %w", err)
	}
	var p struct {
		Iss           string `json:"iss"`
		Sub           string `json:"sub"`
		Aud           any    `json:"aud"`
		Exp           int64  `json:"exp"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(pb, &p); err != nil {
		return nil, fmt.Errorf("google id token: payload parse: %w", err)
	}
	if p.Iss != "https://accounts.google.com" && p.Iss != "accounts.google.com" {
		return nil, fmt.Errorf("google id token: bad iss %q", p.Iss)
	}
	if !googleAudMatches(p.Aud, v.clientID) {
		return nil, errors.New("google id token: aud mismatch")
	}
	if p.Exp == 0 || time.Now().UTC().Unix() >= p.Exp {
		return nil, errors.New("google id token: expired")
	}
	if p.Sub == "" || p.Email == "" {
		return nil, errors.New("google id token: missing sub/email")
	}
	if !p.EmailVerified {
		return nil, errors.New("google id token: email_verified=false")
	}
	return &GoogleClaims{
		Sub:           p.Sub,
		Email:         strings.ToLower(p.Email),
		EmailVerified: p.EmailVerified,
		Name:          p.Name,
		Picture:       p.Picture,
	}, nil
}

func googleAudMatches(aud any, want string) bool {
	switch v := aud.(type) {
	case string:
		return v == want
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == want {
				return true
			}
		}
	}
	return false
}

// googleJWKSCache holds Google's RSA signing keys (one per kid). The
// first miss for a kid triggers a synchronous fetch; entries older than
// `googleJWKSCacheTTL` are treated as stale and re-fetched.
type googleJWKSCache struct {
	url string

	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
}

const googleJWKSCacheTTL = time.Hour

func (c *googleJWKSCache) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if k, ok := c.keys[kid]; ok && time.Since(c.fetched) < googleJWKSCacheTTL {
		c.mu.RUnlock()
		return k, nil
	}
	c.mu.RUnlock()
	if err := c.refresh(ctx); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	k, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("kid %q not found in google jwks", kid)
	}
	return k, nil
}

func (c *googleJWKSCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks fetch: status %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var doc struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return err
	}
	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 | int(b)
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	}
	c.mu.Lock()
	c.keys = keys
	c.fetched = time.Now()
	c.mu.Unlock()
	return nil
}

// googleB64Decode is the unpadded base64url decoder JWT segments use.
// Kept unexported because the verifier parses the JWT by hand (the
// TokenManager's golang-jwt path can't help here — the signing key is
// Google's, not ours).
func googleB64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
