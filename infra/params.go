package infra

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// RequiredQueryInt extracts and parses a required integer query parameter.
// Returns an AppError (bad request) if the parameter is missing or not a valid integer.
func RequiredQueryInt(c fiber.Ctx, name string) (int, error) {
	raw := c.Query(name)
	if raw == "" {
		return 0, ErrBadRequest("missing_"+name, name+" is required")
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, ErrBadRequest("invalid_"+name, name+" must be a number")
	}
	return val, nil
}

// RequiredParamInt extracts and parses a required integer path parameter.
// Returns an AppError (bad request) if the parameter is missing or not a valid integer.
func RequiredParamInt(c fiber.Ctx, name string) (int, error) {
	raw := c.Params(name)
	if raw == "" {
		return 0, ErrBadRequest("missing_"+name, name+" is required")
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, ErrBadRequest("invalid_"+name, name+" must be a number")
	}
	return val, nil
}
