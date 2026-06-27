package httpx

import "github.com/gofiber/fiber/v2"

type AppError struct {
	Status  int
	Message string
}

func (e AppError) Error() string {
	return e.Message
}

func BadRequest(message string) AppError {
	return AppError{Status: fiber.StatusBadRequest, Message: message}
}

func Unauthorized(message string) AppError {
	return AppError{Status: fiber.StatusUnauthorized, Message: message}
}

func Forbidden(message string) AppError {
	return AppError{Status: fiber.StatusForbidden, Message: message}
}

func NotFound(message string) AppError {
	return AppError{Status: fiber.StatusNotFound, Message: message}
}

func Conflict(message string) AppError {
	return AppError{Status: fiber.StatusConflict, Message: message}
}
