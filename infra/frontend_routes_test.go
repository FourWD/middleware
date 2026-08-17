package infra

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// Served from testdata rather than t.TempDir(): fasthttp keeps the served files
// open in its FS cache, which makes TempDir cleanup fail on Windows.
const (
	frontendDist  = "testdata/frontend_dist"
	frontendIndex = "<html>spa</html>"
	frontendAsset = "console.log(1)"
)

func newFrontendApp(t *testing.T, cfg FrontendConfig) *fiber.App {
	t.Helper()
	if cfg.DistPath == "" {
		cfg.DistPath = frontendDist
	}
	app := fiber.New()
	if err := RegisterFrontendRoutes(app, cfg); err != nil {
		t.Fatalf("RegisterFrontendRoutes: %v", err)
	}
	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path string) (int, string) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(method, path, nil))
	if err != nil {
		t.Fatalf("app.Test %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, strings.TrimSpace(string(body))
}

func TestRegisterFrontendRoutes_ErrorsWhenDistPathEmpty(t *testing.T) {
	if err := RegisterFrontendRoutes(fiber.New(), FrontendConfig{}); err == nil {
		t.Fatalf("expected error for empty dist path")
	}
}

func TestRegisterFrontendRoutes_ErrorsWhenIndexMissing(t *testing.T) {
	if err := RegisterFrontendRoutes(fiber.New(), FrontendConfig{DistPath: t.TempDir()}); err == nil {
		t.Fatalf("expected error for missing index.html")
	}
}

func TestRegisterFrontendRoutes_ServesStaticAsset(t *testing.T) {
	status, body := doRequest(t, newFrontendApp(t, FrontendConfig{}), http.MethodGet, "/assets/app.js")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body != frontendAsset {
		t.Fatalf("expected asset content, got %q", body)
	}
}

func TestRegisterFrontendRoutes_DeepLinkFallsBackToIndex(t *testing.T) {
	status, body := doRequest(t, newFrontendApp(t, FrontendConfig{}), http.MethodGet, "/dashboard/orders/42")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body != frontendIndex {
		t.Fatalf("expected index.html, got %q", body)
	}
}

func TestRegisterFrontendRoutes_SkipsAPIPrefix(t *testing.T) {
	status, body := doRequest(t, newFrontendApp(t, FrontendConfig{}), http.MethodGet, "/api/unknown")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
	if body == frontendIndex {
		t.Fatalf("api path must not fall back to index.html")
	}
}

func TestRegisterFrontendRoutes_SkipsNonGETMethods(t *testing.T) {
	status, body := doRequest(t, newFrontendApp(t, FrontendConfig{}), http.MethodPost, "/dashboard")
	if status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
	if body == frontendIndex {
		t.Fatalf("POST must not fall back to index.html")
	}
}

// The fallback is registered after RegisterStandardRoutes, so infra endpoints
// must keep their own responses instead of being shadowed by index.html.
func TestRegisterFrontendRoutes_DoesNotShadowStandardRoutes(t *testing.T) {
	app := fiber.New()
	RegisterStandardRoutes(app, StandardRoutesConfig{})
	if err := RegisterFrontendRoutes(app, FrontendConfig{DistPath: frontendDist}); err != nil {
		t.Fatalf("RegisterFrontendRoutes: %v", err)
	}

	status, body := doRequest(t, app, http.MethodGet, "/livez")
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body == frontendIndex {
		t.Fatalf("/livez must not fall back to index.html")
	}
}

func TestRegisterFrontendRoutes_CustomSkipPrefixes(t *testing.T) {
	app := newFrontendApp(t, FrontendConfig{SkipPrefixes: []string{"/graphql"}})

	if status, _ := doRequest(t, app, http.MethodGet, "/graphql"); status != http.StatusNotFound {
		t.Fatalf("expected 404 for skipped prefix, got %d", status)
	}
	// /api/ is no longer skipped once SkipPrefixes is overridden.
	if _, body := doRequest(t, app, http.MethodGet, "/api/unknown"); body != frontendIndex {
		t.Fatalf("expected index.html, got %q", body)
	}
}
