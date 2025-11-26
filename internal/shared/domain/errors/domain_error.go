package errors

import "fmt"

// ErrorCode representa códigos de erro de domínio
type ErrorCode string

const (
	ErrCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeInternal          ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidState      ErrorCode = "INVALID_STATE"
	ErrCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"
)

// DomainError representa um erro de domínio puro (sem dependências de infraestrutura)
type DomainError struct {
	Code    ErrorCode
	Message string
	Details map[string]string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func (e *DomainError) WithDetail(key, value string) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// Construtores de erros de domínio

func NewValidationError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: make(map[string]string),
	}
}

func NewNotFoundError(resource string) *DomainError {
	return &DomainError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: make(map[string]string),
	}
}

func NewAlreadyExistsError(resource string) *DomainError {
	return &DomainError{
		Code:    ErrCodeAlreadyExists,
		Message: fmt.Sprintf("%s already exists", resource),
		Details: make(map[string]string),
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeUnauthorized,
		Message: message,
		Details: make(map[string]string),
	}
}

func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeForbidden,
		Message: message,
		Details: make(map[string]string),
	}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeConflict,
		Message: message,
		Details: make(map[string]string),
	}
}

func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
		Details: make(map[string]string),
	}
}

func NewInvalidStateError(message string) *DomainError {
	return &DomainError{
		Code:    ErrCodeInvalidState,
		Message: message,
		Details: make(map[string]string),
	}
}

func NewInsufficientFundsError() *DomainError {
	return &DomainError{
		Code:    ErrCodeInsufficientFunds,
		Message: "insufficient funds for this operation",
		Details: make(map[string]string),
	}
}
