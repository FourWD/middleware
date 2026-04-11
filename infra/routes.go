package infra

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
)

// StandardRoutesConfig configures the standard routes that every project should have.
type StandardRoutesConfig struct {
	// AppEnv is the validated application environment, e.g. local/dev/prod/test.
	AppEnv string

	// HealthReport is called by /healthz and /readyz to determine service readiness.
	HealthReport func() HealthReport

	// HeartbeatDebugStatus returns the heartbeat debug payload for
	// /debug/background-jobs/heartbeat/circuit-state.
	// If nil, the endpoint returns {"enabled": false}.
	HeartbeatDebugStatus func() any

	// DebugAuthToken protects debug routes when non-empty.
	// Clients must send it via X-Debug-Auth header.
	DebugAuthToken string
}

// RegisterStandardRoutes registers /healthz, /metrics, and /debug endpoints.
func RegisterStandardRoutes(app *fiber.App, cfg StandardRoutesConfig) {
	appEnv := strings.ToLower(strings.TrimSpace(cfg.AppEnv))
	if appEnv == "" {
		appEnv = "local"
	}
	debugRouteEnabled := appEnv != "prod" && strings.TrimSpace(cfg.DebugAuthToken) != ""

	app.Get("/livez", func(c fiber.Ctx) error {
		return WriteSuccess(c, fiber.StatusOK, fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/healthz", func(c fiber.Ctx) error {
		report := HealthReport{Status: "ok"}
		if cfg.HealthReport != nil {
			report = cfg.HealthReport()
		}
		status := fiber.StatusOK
		if !report.Healthy() {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(Envelope{
			Success: report.Healthy(),
			Data:    report,
		})
	})

	app.Get("/readyz", func(c fiber.Ctx) error {
		report := HealthReport{Status: "ok"}
		if cfg.HealthReport != nil {
			report = cfg.HealthReport()
		}
		status := fiber.StatusOK
		if !report.Healthy() {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(Envelope{
			Success: report.Healthy(),
			Data:    report,
		})
	})

	app.Get("/metrics", adaptor.HTTPHandler(PrometheusHandler()))

	if debugRouteEnabled {
		app.Get("/debug/background-jobs/heartbeat/circuit-state", func(c fiber.Ctx) error {
			debugToken := c.Get("X-Debug-Auth")
			if subtle.ConstantTimeCompare([]byte(debugToken), []byte(cfg.DebugAuthToken)) != 1 {
				return WriteErrorEnvelope(c, fiber.StatusUnauthorized, "unauthorized", "invalid debug auth token")
			}

			payload := any(fiber.Map{
				"enabled": false,
			})
			if cfg.HeartbeatDebugStatus != nil {
				payload = cfg.HeartbeatDebugStatus()
			}

			return WriteSuccess(c, fiber.StatusOK, fiber.Map{
				"heartbeat": payload,
			})
		})
	}
}
