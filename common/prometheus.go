package common

import (
	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

var prometheusMiddleware fiber.Handler

func RegisterPrometheus(app *fiber.App, name string) {
	p := fiberprometheus.New(name)
	p.RegisterAt(app, "/metrics")

	prometheusMiddleware = p.Middleware
}

func FiberPrometheus(c *fiber.Ctx) error {
	if prometheusMiddleware == nil {
		return c.Next()
	}
	return prometheusMiddleware(c)
}
