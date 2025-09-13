package server

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"
	errorCode := "INTERNAL_ERROR"

	// Handle Fiber errors
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
		
		switch code {
		case fiber.StatusBadRequest:
			errorCode = "BAD_REQUEST"
		case fiber.StatusUnauthorized:
			errorCode = "UNAUTHORIZED"
		case fiber.StatusForbidden:
			errorCode = "FORBIDDEN"
		case fiber.StatusNotFound:
			errorCode = "NOT_FOUND"
		case fiber.StatusTooManyRequests:
			errorCode = "RATE_LIMIT_EXCEEDED"
		case fiber.StatusInternalServerError:
			errorCode = "INTERNAL_ERROR"
		}
	}

	// Log the error
	slog.Error("Request error",
		"error", err.Error(),
		"code", code,
		"path", c.Path(),
		"method", c.Method(),
		"ip", c.IP(),
	)

	// Return error response
	return c.Status(code).JSON(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

func SuccessResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func PaginatedResponse(c *fiber.Ctx, data interface{}, total int, page, limit int) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"pagination": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func ValidationErrorResponse(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: errors,
	})
}