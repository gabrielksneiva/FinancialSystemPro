package api

import (
	"financial-system-pro/domain"
	"financial-system-pro/internal/validator"

	"github.com/gofiber/fiber/v2"
)

var validatorService = validator.New()

// ValidateRequest valida o request body e retorna AppError se inválido
func ValidateRequest(ctx *fiber.Ctx, body interface{}) *domain.AppError {
	err := ctx.BodyParser(body)
	if err != nil {
		return domain.NewValidationError("body", "Invalid JSON format")
	}

	return validatorService.Validate(body)
}
