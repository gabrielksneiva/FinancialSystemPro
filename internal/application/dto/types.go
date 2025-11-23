package dto

import (
	"financial-system-pro/internal/domain/errors"
	"financial-system-pro/internal/shared/validator"

	"github.com/gofiber/fiber/v2"
)

// ValidateRequest valida o request body e retorna AppError se inválido
func ValidateRequest(ctx *fiber.Ctx, body interface{}) *errors.AppError {
	err := ctx.BodyParser(body)
	if err != nil {
		return errors.NewValidationError("body", "Invalid JSON format")
	}

	validatorService := validator.New()
	return validatorService.Validate(body)
}
