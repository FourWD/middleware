package infra

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// --- composeExemptSet -------------------------------------------------------

func TestComposeExemptSet_IncludesBaselineByDefault(t *testing.T) {
	t.Setenv("HTTP_PUBLIC_PATHS", "")
	got := composeExemptSet(AuthMiddlewareV2Config{})

	for _, want := range baselineExemptPaths {
		if _, ok := got[want]; !ok {
			t.Errorf("baseline path %q missing", want)
		}
	}
}

func TestComposeExemptSet_IncludesEnvPublicPaths(t *testing.T) {
	t.Setenv("HTTP_PUBLIC_PATHS", "/api/v1/auth/login,/api/v1/auth/refresh")
	got := composeExemptSet(AuthMiddlewareV2Config{})

	if _, ok := got["/api/v1/auth/login"]; !ok {
		t.Error("/api/v1/auth/login from env missing")
	}
	if _, ok := got["/api/v1/auth/refresh"]; !ok {
		t.Error("/api/v1/auth/refresh from env missing")
	}
}

func TestComposeExemptSet_IncludesCallerPaths(t *testing.T) {
	t.Setenv("HTTP_PUBLIC_PATHS", "")
	got := composeExemptSet(AuthMiddlewareV2Config{
		ExemptPaths: []string{"/custom/health"},
	})

	if _, ok := got["/custom/health"]; !ok {
		t.Error("/custom/health from caller missing")
	}
}

func TestComposeExemptSet_DisableAutoExempts(t *testing.T) {
	t.Setenv("HTTP_PUBLIC_PATHS", "/env-path")
	got := composeExemptSet(AuthMiddlewareV2Config{
		DisableAutoExempts: true,
		ExemptPaths:        []string{"/only-this"},
	})

	if _, ok := got["/only-this"]; !ok {
		t.Error("caller path missing")
	}
	for _, baseline := range baselineExemptPaths {
		if _, ok := got[baseline]; ok {
			t.Errorf("baseline %q should be excluded when DisableAutoExempts=true", baseline)
		}
	}
	if _, ok := got["/env-path"]; ok {
		t.Error("env path should be excluded when DisableAutoExempts=true")
	}
}

func TestComposeExemptSet_DuplicatesCollapse(t *testing.T) {
	t.Setenv("HTTP_PUBLIC_PATHS", "/metrics") // duplicates a baseline
	got := composeExemptSet(AuthMiddlewareV2Config{
		ExemptPaths: []string{"/metrics", "/metrics"}, // dup in caller too
	})

	// Set semantics — only one entry regardless of how many sources added it
	if _, ok := got["/metrics"]; !ok {
		t.Error("/metrics missing")
	}
}

// --- AuthSession context plumbing -------------------------------------------

func TestAuthSessionFrom_NoSessionReturnsFalse(t *testing.T) {
	_, ok := AuthSessionFrom(context.Background())
	if ok {
		t.Fatal("expected ok=false for ctx without session")
	}
}

func TestAuthSessionFrom_NilContext(t *testing.T) {
	_, ok := AuthSessionFrom(nil)
	if ok {
		t.Fatal("expected ok=false for nil ctx")
	}
}

func TestWithAuthSession_RoundTrip(t *testing.T) {
	want := AuthSession{UserID: "u1", Email: "alice@example.com", Role: "admin"}
	ctx := WithAuthSession(context.Background(), want)

	got, ok := AuthSessionFrom(ctx)
	if !ok {
		t.Fatal("expected ok=true after WithAuthSession")
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

// --- extractBearerToken via fiber harness -----------------------------------

func runExtract(t *testing.T, authHeader string) (status int) {
	t.Helper()
	app := fiber.New()
	app.Get("/extract", func(c fiber.Ctx) error {
		// extractBearerToken writes a 401 response itself on failure and
		// returns ("", nil) — there is no separate error to propagate.
		// Token=="" is the failure signal; success returns the token.
		token, _ := extractBearerToken(c)
		if token == "" {
			return nil // response already written
		}
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/extract", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp.StatusCode
}

func TestExtractBearerToken_NoHeader(t *testing.T) {
	if got := runExtract(t, ""); got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", got)
	}
}

func TestExtractBearerToken_WrongScheme(t *testing.T) {
	if got := runExtract(t, "Basic dXNlcjpwYXNz"); got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", got)
	}
}

func TestExtractBearerToken_EmptyToken(t *testing.T) {
	if got := runExtract(t, "Bearer "); got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", got)
	}
}

func TestExtractBearerToken_WhitespaceOnlyToken(t *testing.T) {
	if got := runExtract(t, "Bearer    "); got != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", got)
	}
}

func TestExtractBearerToken_ValidToken(t *testing.T) {
	if got := runExtract(t, "Bearer abc.def.ghi"); got != http.StatusOK {
		t.Fatalf("got %d want 200", got)
	}
}

// --- AdminOnly --------------------------------------------------------------

func TestAdminOnly_NoSession_Returns401(t *testing.T) {
	app := fiber.New()
	app.Get("/admin", AdminOnly("admin", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", resp.StatusCode)
	}
}

func TestAdminOnly_WrongRole_Returns403(t *testing.T) {
	app := fiber.New()
	// Inject a non-admin session into ctx via test middleware
	app.Use(func(c fiber.Ctx) error {
		c.SetContext(WithAuthSession(c.Context(), AuthSession{UserID: "u1", Role: "user"}))
		return c.Next()
	})
	app.Get("/admin", AdminOnly("admin", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("got %d want 403", resp.StatusCode)
	}
}

func TestAdminOnly_CorrectRole_Passes(t *testing.T) {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.SetContext(WithAuthSession(c.Context(), AuthSession{UserID: "u1", Role: "admin"}))
		return c.Next()
	})
	app.Get("/admin", AdminOnly("admin", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d want 200", resp.StatusCode)
	}
}
