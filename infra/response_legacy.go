package infra

import "github.com/gofiber/fiber/v3"

// LegacyEnvelope matches the old FourWD/middleware/common response structure.
// { status: 1|0, code: string, message: string, data: any }
//
// Deprecated: use Envelope and WriteSuccess/WriteErrorEnvelope from response.go for new projects.
// This file exists only for backward compatibility with existing FourWD/middleware/common clients.
// Do not use WriteLegacy* in new projects that do not need to match the legacy contract.
type LegacyEnvelope struct {
	Status  int    `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteLegacySuccess outputs a success response in the legacy JSON format.
// It maps the "message" key from a fiber.Map (if it exists) to the top-level Message.
func WriteLegacySuccess(c fiber.Ctx, status int, data any) error {
	var msg string
	if m, ok := data.(fiber.Map); ok {
		if val, exists := m["message"]; exists {
			if s, ok := val.(string); ok {
				msg = s
				delete(m, "message")
				if len(m) == 0 {
					data = nil
				}
			}
		}
	}

	env := LegacyEnvelope{
		Status: 1,
		Code:   "0000",
	}

	if msg != "" {
		env.Message = msg
	} else {
		env.Message = "success"
	}

	env.Data = data

	return c.Status(status).JSON(env)
}

// WriteLegacyErrorEnvelope outputs an error response in the legacy JSON format.
func WriteLegacyErrorEnvelope(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(LegacyEnvelope{
		Status:  0,
		Code:    code,
		Message: message,
	})
}

// WriteLegacyError converts an error into a standard legacy JSON error response.
func WriteLegacyError(c fiber.Ctx, status int, err error) error {
	return WriteLegacyErrorEnvelope(c, status, "1001", err.Error())
}
