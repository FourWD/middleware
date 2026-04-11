package infra

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRegisterStandardRoutes_LivezAlwaysOK(t *testing.T) {
	app := fiber.New()
	RegisterStandardRoutes(app, StandardRoutesConfig{})

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRegisterStandardRoutes_HealthzReportsFailure(t *testing.T) {
	app := fiber.New()
	RegisterStandardRoutes(app, StandardRoutesConfig{
		HealthReport: func() HealthReport {
			return HealthReport{
				Status: "down",
				Components: map[string]HealthComponentStatus{
					"database": {Status: "down", Required: true, Error: "dial tcp timeout"},
				},
			}
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}

	var body Envelope
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success {
		t.Fatalf("expected success=false")
	}
}

func TestRegisterStandardRoutes_DebugRouteDisabledInProd(t *testing.T) {
	app := fiber.New()
	RegisterStandardRoutes(app, StandardRoutesConfig{
		AppEnv:         "prod",
		DebugAuthToken: "1234567890abcdef",
	})

	req := httptest.NewRequest(http.MethodGet, "/debug/background-jobs/heartbeat/circuit-state", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRegisterStandardRoutes_DebugRouteRequiresToken(t *testing.T) {
	app := fiber.New()
	RegisterStandardRoutes(app, StandardRoutesConfig{
		AppEnv:         "dev",
		DebugAuthToken: "1234567890abcdef",
	})

	reqNoToken := httptest.NewRequest(http.MethodGet, "/debug/background-jobs/heartbeat/circuit-state", nil)
	respNoToken, err := app.Test(reqNoToken)
	if err != nil {
		t.Fatalf("app.Test without token failed: %v", err)
	}
	if respNoToken.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", respNoToken.StatusCode)
	}

	reqWithToken := httptest.NewRequest(http.MethodGet, "/debug/background-jobs/heartbeat/circuit-state", nil)
	reqWithToken.Header.Set("X-Debug-Auth", "1234567890abcdef")
	respWithToken, err := app.Test(reqWithToken)
	if err != nil {
		t.Fatalf("app.Test with token failed: %v", err)
	}
	if respWithToken.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 with token, got %d", respWithToken.StatusCode)
	}
}

func TestRegisterStandardRoutes_DebugRouteDisabledWithoutToken(t *testing.T) {
	app := fiber.New()
	RegisterStandardRoutes(app, StandardRoutesConfig{
		AppEnv: "dev",
	})

	req := httptest.NewRequest(http.MethodGet, "/debug/background-jobs/heartbeat/circuit-state", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
