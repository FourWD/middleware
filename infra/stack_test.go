package infra

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRegisterCORSPermitsDefaultUploadHeaders(t *testing.T) {
	app := fiber.New()
	RegisterBaseStack(app, StackConfig{AllowOrigins: "*"})
	app.Post("/api/v1/upload/", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/upload/", nil)
	req.Header.Set("Origin", "https://backoffice.omakasecarsauction.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type,authorization,x-content-type-options")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}

	allowHeaders := strings.ToLower(resp.Header.Get("Access-Control-Allow-Headers"))
	for _, want := range []string{"content-type", "authorization", "x-content-type-options"} {
		if !strings.Contains(allowHeaders, want) {
			t.Fatalf("Access-Control-Allow-Headers = %q, missing %q", allowHeaders, want)
		}
	}
}

func TestLoadStackConfigReadsHTTPAllowHeaders(t *testing.T) {
	t.Setenv("HTTP_ALLOW_HEADERS", "Content-Type,X-Custom")

	cfg := LoadStackConfig()

	if cfg.AllowHeaders != "Content-Type,X-Custom" {
		t.Fatalf("AllowHeaders = %q, want %q", cfg.AllowHeaders, "Content-Type,X-Custom")
	}
}
