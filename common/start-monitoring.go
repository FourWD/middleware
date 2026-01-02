package common

import "github.com/gofiber/fiber/v2"

func StartMonitoring(name string, app *fiber.App, logic interface{}) {
	runLatestVersionOnly()
	registerPrometheus(name, app, logic)
	monitorDatabaseConnectionPool()
}
