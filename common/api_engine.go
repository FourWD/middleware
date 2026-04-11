package common

// ===== จะย้ายไป middleware ทีหลัง =====

import (
	"github.com/gofiber/fiber/v2"
)

var fiberApp *fiber.App
var serverErrChan chan error

func APIEngine(app *fiber.App, port string) {
	fiberApp = app
	fiberWarmUp(fiberApp)
	fiberWakeUp(fiberApp)
	registerPrometheus()
	serverErrChan = make(chan error, 1)

	go func() {
		if err := fiberApp.Listen(":" + port); err != nil {
			serverErrChan <- err
		}
	}()
}
