package infra

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func preflightAllowHeaders(t *testing.T, cfg StackConfig) string {
	t.Helper()

	app := fiber.New()
	RegisterBaseStack(app, cfg)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/upload/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type,authorization")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}

	return strings.ToLower(resp.Header.Get("Access-Control-Allow-Headers"))
}

func TestRegisterCORSPermitsDefaultHeaders(t *testing.T) {
	got := preflightAllowHeaders(t, StackConfig{AllowOrigins: "*"})

	for _, want := range []string{"origin", "content-type", "accept", "authorization", "x-request-id"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Access-Control-Allow-Headers = %q, missing %q", got, want)
		}
	}
}

// The default list stays minimal — client-specific headers go in
// HTTP_ALLOW_HEADERS_EXTRA, not here.
func TestAllowHeadersDefaultsAreMinimal(t *testing.T) {
	want := []string{"Origin", "Content-Type", "Accept", "Authorization", RequestIDHeader}

	got := allowHeaders("")

	if len(got) != len(want) {
		t.Fatalf("allowHeaders(\"\") = %v (%d headers), want %v (%d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("header %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAllowHeadersExtraCarriesClientWorkarounds(t *testing.T) {
	got := preflightAllowHeaders(t, StackConfig{
		AllowOrigins:      "*",
		AllowHeadersExtra: "X-Content-Type-Options,X-Requested-With",
	})

	for _, want := range []string{"authorization", "x-content-type-options", "x-requested-with"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Access-Control-Allow-Headers = %q, missing %q", got, want)
		}
	}
}

// Extras must never drop a default — a deployment that sets only X-Tenant-ID
// must still allow Authorization, otherwise every browser request breaks.
func TestRegisterCORSExtraHeadersAddToDefaults(t *testing.T) {
	got := preflightAllowHeaders(t, StackConfig{AllowOrigins: "*", AllowHeadersExtra: "X-Tenant-ID"})

	for _, want := range []string{"authorization", "content-type", "x-tenant-id"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Access-Control-Allow-Headers = %q, missing %q", got, want)
		}
	}
}

func TestAllowHeadersDeduplicatesCaseInsensitively(t *testing.T) {
	got := allowHeaders("authorization, X-REQUEST-ID ,X-Tenant-ID")

	seen := map[string]int{}
	for _, h := range got {
		seen[strings.ToLower(h)]++
	}
	for h, n := range seen {
		if n > 1 {
			t.Fatalf("header %q appears %d times in %v", h, n, got)
		}
	}
	if seen["x-tenant-id"] != 1 {
		t.Fatalf("extra header missing from %v", got)
	}
}

func TestRegisterCORSSetsMaxAge(t *testing.T) {
	app := fiber.New()
	RegisterBaseStack(app, StackConfig{AllowOrigins: "*", CORSMaxAge: 600})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/upload/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("Access-Control-Max-Age = %q, want %q", got, "600")
	}
}

func TestLoadStackConfigReadsCORSEnv(t *testing.T) {
	t.Setenv("HTTP_ALLOW_HEADERS_EXTRA", "X-Custom")
	t.Setenv("HTTP_CORS_MAX_AGE", "120")

	cfg := LoadStackConfig()

	if cfg.AllowHeadersExtra != "X-Custom" {
		t.Fatalf("AllowHeadersExtra = %q, want %q", cfg.AllowHeadersExtra, "X-Custom")
	}
	if cfg.CORSMaxAge != 120 {
		t.Fatalf("CORSMaxAge = %d, want 120", cfg.CORSMaxAge)
	}
}

func TestLoadStackConfigCORSDefaults(t *testing.T) {
	t.Setenv("HTTP_ALLOW_HEADERS_EXTRA", "")
	t.Setenv("HTTP_CORS_MAX_AGE", "")

	cfg := LoadStackConfig()

	if cfg.AllowHeadersExtra != "" {
		t.Fatalf("AllowHeadersExtra = %q, want empty", cfg.AllowHeadersExtra)
	}
	if cfg.CORSMaxAge != 600 {
		t.Fatalf("CORSMaxAge = %d, want 600", cfg.CORSMaxAge)
	}
}
