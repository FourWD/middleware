package infra

import "github.com/gofiber/fiber/v3"

// AppError is a domain error that carries an HTTP status code and machine-readable code.
// Handlers can return AppError and the global error handler maps it automatically.
type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.Cause }

func NewAppError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func ErrBadRequest(code, message string) *AppError {
	return NewAppError(fiber.StatusBadRequest, code, message)
}

// ErrBadRequestWrap creates a bad request error wrapping the cause for error chain preservation.
func ErrBadRequestWrap(code string, cause error) *AppError {
	return &AppError{Status: fiber.StatusBadRequest, Code: code, Message: cause.Error(), Cause: cause}
}

func ErrNotFound(code, message string) *AppError {
	return NewAppError(fiber.StatusNotFound, code, message)
}

func ErrConflict(code, message string) *AppError {
	return NewAppError(fiber.StatusConflict, code, message)
}
