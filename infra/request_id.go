package infra

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

func NewRequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		SetRequestID(c, requestID)
		c.Set(RequestIDHeader, requestID)
		c.SetContext(InjectCorrelationID(c.Context(), requestID))

		return c.Next()
	}
}
