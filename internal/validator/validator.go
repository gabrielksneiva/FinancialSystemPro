package validator

import (
	"financial-system-pro/domain"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidatorService encapsula o validador
type ValidatorService struct {
	validate *validator.Validate
}

// New cria uma nova instância do validador
func New() *ValidatorService {
	return &ValidatorService{
		validate: validator.New(),
	}
}

// Validate valida um struct e retorna AppError se houver erro
func (v *ValidatorService) Validate(data interface{}) *domain.AppError {
	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	// Erro interno de validação
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return domain.NewInternalError("Invalid validation error", nil)
	}

	// Erros de validação de campo
	validationErrors := err.(validator.ValidationErrors)
	if len(validationErrors) > 0 {
		firstErr := validationErrors[0]
		field := firstErr.Field()
		tag := firstErr.Tag()
		message := v.getErrorMessage(field, tag, firstErr.Param())
		return domain.NewValidationError(field, message)
	}

	return domain.NewValidationError("unknown", "Validation failed")
}

// getErrorMessage retorna uma mensagem amigável baseada na tag de validação
func (v *ValidatorService) getErrorMessage(field, tag, param string) string {
	messages := map[string]string{
		"required": fmt.Sprintf("%s is required", field),
		"email":    fmt.Sprintf("%s must be a valid email", field),
		"min":      fmt.Sprintf("%s must be at least %s characters", field, param),
		"max":      fmt.Sprintf("%s must be at most %s characters", field, param),
		"len":      fmt.Sprintf("%s must have exactly %s characters", field, param),
		"numeric":  fmt.Sprintf("%s must be numeric", field),
		"url":      fmt.Sprintf("%s must be a valid URL", field),
		"gt":       fmt.Sprintf("%s must be greater than %s", field, param),
		"gte":      fmt.Sprintf("%s must be greater than or equal to %s", field, param),
		"lt":       fmt.Sprintf("%s must be less than %s", field, param),
		"lte":      fmt.Sprintf("%s must be less than or equal to %s", field, param),
		"eq":       fmt.Sprintf("%s must be equal to %s", field, param),
		"ne":       fmt.Sprintf("%s must not be equal to %s", field, param),
	}

	if msg, exists := messages[tag]; exists {
		return msg
	}

	return fmt.Sprintf("%s validation failed on tag '%s'", field, tag)
}
