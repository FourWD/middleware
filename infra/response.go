package infra

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteSuccess(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(Envelope{
		Success: true,
		Data:    data,
	})
}

func WriteErrorEnvelope(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// WriteError converts an error into a standard JSON error response.
// If the error is an AppError, its code and message are used directly.
// Otherwise, a generic message is returned to avoid leaking internal details.
func WriteError(c fiber.Ctx, status int, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return WriteErrorEnvelope(c, status, appErr.Code, appErr.Message)
	}

	code := "internal_error"
	message := "an unexpected error occurred"
	if status >= 400 && status < 500 {
		code = "bad_request"
		message = "the request could not be processed"
	}
	return WriteErrorEnvelope(c, status, code, message)
}
