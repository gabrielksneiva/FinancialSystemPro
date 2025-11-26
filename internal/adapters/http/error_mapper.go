package http

import (
	"net/http"

	sharedErrors "financial-system-pro/internal/shared/domain/errors"
)

// MapDomainErrorToHTTP mapeia erros de domínio para códigos de status HTTP
func MapDomainErrorToHTTP(err error) int {
	if domainErr, ok := err.(*sharedErrors.DomainError); ok {
		switch domainErr.Code {
		case sharedErrors.ErrCodeValidation:
			return http.StatusBadRequest
		case sharedErrors.ErrCodeNotFound:
			return http.StatusNotFound
		case sharedErrors.ErrCodeAlreadyExists:
			return http.StatusConflict
		case sharedErrors.ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case sharedErrors.ErrCodeForbidden:
			return http.StatusForbidden
		case sharedErrors.ErrCodeConflict:
			return http.StatusConflict
		case sharedErrors.ErrCodeInvalidState:
			return http.StatusBadRequest
		case sharedErrors.ErrCodeInsufficientFunds:
			return http.StatusBadRequest
		case sharedErrors.ErrCodeInternal:
			return http.StatusInternalServerError
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// ErrorResponse representa a resposta de erro HTTP
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// MapDomainErrorToResponse mapeia erro de domínio para resposta HTTP
func MapDomainErrorToResponse(err error) ErrorResponse {
	if domainErr, ok := err.(*sharedErrors.DomainError); ok {
		return ErrorResponse{
			Code:    string(domainErr.Code),
			Message: domainErr.Message,
			Details: domainErr.Details,
		}
	}
	return ErrorResponse{
		Code:    string(sharedErrors.ErrCodeInternal),
		Message: "an unexpected error occurred",
	}
}
