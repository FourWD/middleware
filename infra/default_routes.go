package infra

import (
	"github.com/gofiber/fiber/v3"
)

// registerGAERoutes mounts GAE lifecycle endpoints.
//
//   - GET /_ah/warmup : warmup ping sent by App Engine before routing traffic.
//     Returns 200 so the instance is considered healthy.
//   - GET /wake-up    : returns the current service version; consumed by the
//     GAE version watcher to detect new deployments.
//
// Mounted automatically from NewApp so services never have to register them manually.
func registerDefaultRoutes(app *fiber.App, cfg CommonConfig) {
	app.Get("/_ah/warmup", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  1,
			"code":    "0000",
			"message": "warm-up request succeeded",
		})
	})

	app.Get("/wake-up", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  1,
			"message": "success",
			"data": fiber.Map{
				"app_id":      cfg.AppID,
				"app_version": cfg.AppVersion,
				"env":         cfg.AppEnv,
			},
		})
	})
}
