package common

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

func FiberMetrics(app *fiber.App, name string) {
	prometheus := fiberprometheus.New(name)
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
}
