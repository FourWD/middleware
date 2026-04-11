package infra

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// NewEnvelopeWrapper wraps successful JSON responses in the standard Envelope format.
// Responses that are already error envelopes (success=false) or non-JSON are passed through unchanged.
func NewEnvelopeWrapper() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			return err
		}

		status := c.Response().StatusCode()

		// No-content responses don't need wrapping
		if status == fiber.StatusNoContent || len(c.Response().Body()) == 0 {
			return nil
		}

		// Only wrap JSON responses
		contentType := string(c.Response().Header.ContentType())
		if len(contentType) < 16 || contentType[:16] != "application/json" {
			return nil
		}

		body := c.Response().Body()

		// Check if already wrapped in an Envelope (has "success" key at top level)
		var probe map[string]json.RawMessage
		if json.Unmarshal(body, &probe) == nil {
			if _, hasSuccess := probe["success"]; hasSuccess {
				return nil
			}
		}

		// Parse the original body as generic JSON
		var data any
		if err := json.Unmarshal(body, &data); err != nil {
			return nil
		}

		wrapped := Envelope{
			Success: true,
			Data:    data,
		}
		out, err := json.Marshal(wrapped)
		if err != nil {
			return nil
		}

		c.Response().SetBody(out)
		return nil
	}
}
