package infra

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// FiberCustom writes a JSON response with the given HTTP status. Always
// injects the request_id into the response body.
func FiberCustom(c fiber.Ctx, HTTPStatus int, data map[string]interface{}) error {
	c.Set("Content-Type", "application/json")
	data["request_id"] = GetRequestID(c)
	return c.Status(HTTPStatus).JSON(data)
}

// FiberOK writes a 200 response with the standard {status, code, message}
// envelope.
func FiberOK(c fiber.Ctx, status int, code string, message string) error {
	response := map[string]interface{}{
		"status":  status,
		"code":    code,
		"message": message,
	}
	return FiberCustom(c, fiber.StatusOK, response)
}

// FiberSuccess writes a 200 response with status=1, code="0000".
func FiberSuccess(c fiber.Ctx) error {
	return FiberOK(c, 1, "0000", "success")
}

// FiberError writes a 500 response with the error envelope and logs the
// failure. If err is provided, logs at Error level; otherwise Warn.
func FiberError(c fiber.Ctx, code string, message string, err ...error) error {
	data := map[string]any{"code": code, "message": message}
	if len(err) > 0 && err[0] != nil {
		AppLog.ErrorCtx(c.Context(), err[0], "FIBER_HANDLER_FAILURE", data,
			WithComponent(ComponentHandler),
			WithOperation("write_error"),
			WithLogKind(LogKindError))
	} else {
		AppLog.WarnCtx(c.Context(), "FIBER_HANDLER_FAILURE", data,
			WithComponent(ComponentHandler),
			WithOperation("write_error"),
			WithLogKind(LogKindError))
	}

	response := map[string]interface{}{
		"status":  0,
		"code":    code,
		"message": message,
	}
	return FiberCustom(c, fiber.StatusInternalServerError, response)
}

// FiberReviewPayload writes a 500 response with code="1002", "review your payload".
func FiberReviewPayload(c fiber.Ctx) error {
	return FiberError(c, "1002", "review your payload")
}

// FiberSendData writes a 200 response with the JSON payload as raw data.
// In non-prod environments, the source SQL is also included for debugging.
func FiberSendData(c fiber.Ctx, jsonData string, sql string) error {
	response := map[string]interface{}{
		"status":  1,
		"message": "success",
		"data":    json.RawMessage(jsonData),
	}

	if AppInfo.Env != "prod" {
		response["sql"] = sql
	}

	return FiberCustom(c, fiber.StatusOK, response)
}
