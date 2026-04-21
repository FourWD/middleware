package infra

import "github.com/gofiber/fiber/v3"

// FiberDisableXFrame sets X-Frame-Options: DENY on every response to prevent
// clickjacking via <iframe> embedding on other origins.
//
// Attach manually in your Register function when needed:
//
//	app.Use(infra.FiberDisableXFrame)
func FiberDisableXFrame(c fiber.Ctx) error {
	c.Set("X-Frame-Options", "DENY")
	return c.Next()
}

// FiberNoSniff sets X-Content-Type-Options: nosniff on every response to
// prevent browsers from MIME-sniffing a response away from the declared
// Content-Type.
//
// Attach manually in your Register function when needed:
//
//	app.Use(infra.FiberNoSniff)
func FiberNoSniff(c fiber.Ctx) error {
	c.Set("X-Content-Type-Options", "nosniff")
	return c.Next()
}
