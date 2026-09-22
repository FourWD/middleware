package infra

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// responseStatus is the status the client receives, read by middleware
// right after c.Next().
//
// When a handler returns an error (an *AppError, fiber.NewError, or the
// router's ErrNotFound for an unmatched path), fiber writes the status later,
// in the app ErrorHandler — after every middleware has already returned. At
// this point c.Response().StatusCode() still reads the default 200, which is
// how every 404 scan logged "GET / -> 200" and was counted as 2xx. This
// mirrors AppErrorHandler's mapping so the value matches what the client
// actually receives.
func responseStatus(c fiber.Ctx, err error) int {
	if err == nil {
		return c.Response().StatusCode()
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code
	}
	return fiber.StatusInternalServerError
}

func routePath(c fiber.Ctx) string {
	if fullPath := c.FullPath(); fullPath != "" {
		return fullPath
	}

	return c.Path()
}

func statusCodeClass(status int) string {
	if status < 100 || status > 999 {
		return "unknown"
	}
	return strconv.Itoa(status/100) + "xx"
}
