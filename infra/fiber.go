package infra

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

type FiberConfig struct {
	AppID       string
	ProxyHeader string
}

func NewFiberApp(cfg FiberConfig) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:         cfg.AppID,
		ProxyHeader:     cfg.ProxyHeader,
		ErrorHandler:    AppErrorHandler(),
		StructValidator: NewValidator(),
	})
}

func AppErrorHandler() func(fiber.Ctx, error) error {
	return func(c fiber.Ctx, err error) error {
		// Domain errors with explicit status/code
		var appErr *AppError
		if errors.As(err, &appErr) {
			return WriteErrorEnvelope(c, appErr.Status, appErr.Code, appErr.Message)
		}

		// Fiber errors (e.g. JSON bind failures, 404 from router)
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return WriteErrorEnvelope(c, fiberErr.Code, "request_error", fiberErr.Message)
		}

		// Unknown errors — hide details from client
		return WriteError(c, fiber.StatusInternalServerError, err)
	}
}
