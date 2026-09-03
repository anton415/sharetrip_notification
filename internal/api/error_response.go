package api

import "github.com/gofiber/fiber/v2"

const (
	errorCodeValidation = "VALIDATION_ERROR"
	errorCodeNotFound   = "NOT_FOUND"
	errorCodeInternal   = "INTERNAL_ERROR"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeErrorResponse(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(errorResponse{
		Code:    code,
		Message: message,
	})
}
