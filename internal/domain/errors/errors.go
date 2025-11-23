package errors

import "net/http"

// Error codes
const (
	// Auth errors
	ErrInvalidPassword = "INVALID_PASSWORD"
	ErrUserNotFound    = "USER_NOT_FOUND"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrTokenExpired    = "TOKEN_EXPIRED"
	ErrInvalidToken    = "INVALID_TOKEN"

	// Validation errors
	ErrValidation    = "VALIDATION_ERROR"
	ErrInvalidEmail  = "INVALID_EMAIL"
	ErrInvalidAmount = "INVALID_AMOUNT"
	ErrMissingField  = "MISSING_FIELD"

	// Database errors
	ErrDatabaseConnection = "DATABASE_CONNECTION_ERROR"
	ErrDatabaseQuery      = "DATABASE_QUERY_ERROR"
	ErrRecordNotFound     = "RECORD_NOT_FOUND"
	ErrDuplicateRecord    = "DUPLICATE_RECORD"

	// Tron errors
	ErrTronConnection    = "TRON_CONNECTION_ERROR"
	ErrTronRPC           = "TRON_RPC_ERROR"
	ErrInvalidWallet     = "INVALID_WALLET_ADDRESS"
	ErrInsufficientFunds = "INSUFFICIENT_FUNDS"

	// Server errors
	ErrInternal       = "INTERNAL_SERVER_ERROR"
	ErrNotImplemented = "NOT_IMPLEMENTED"
)

// AppError é o tipo de erro padrão da aplicação
type AppError struct {
	Details    map[string]interface{} `json:"details,omitempty"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
}

// Error implementa a interface error
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError cria um novo erro de aplicação
func NewAppError(code, message string, statusCode int, details map[string]interface{}) *AppError {
	if details == nil {
		details = make(map[string]interface{})
	}
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

// NewValidationError cria erro de validação
func NewValidationError(field, message string) *AppError {
	return NewAppError(
		ErrValidation,
		"Validation error: "+message,
		http.StatusBadRequest,
		map[string]interface{}{"field": field},
	)
}

// NewUnauthorizedError cria erro de autenticação
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(
		ErrUnauthorized,
		message,
		http.StatusUnauthorized,
		nil,
	)
}

// NewInternalError cria erro interno do servidor
func NewInternalError(message string, details map[string]interface{}) *AppError {
	return NewAppError(
		ErrInternal,
		message,
		http.StatusInternalServerError,
		details,
	)
}

// NewDatabaseError cria erro de banco de dados
func NewDatabaseError(operation string, err error) *AppError {
	message := "Database error"
	if err != nil {
		message += ": " + err.Error()
	}
	return NewAppError(
		ErrDatabaseConnection,
		message,
		http.StatusInternalServerError,
		map[string]interface{}{"operation": operation},
	)
}

// NewNotFoundError cria erro de recurso não encontrado
func NewNotFoundError(resource string) *AppError {
	return NewAppError(
		ErrRecordNotFound,
		resource+" not found",
		http.StatusNotFound,
		map[string]interface{}{"resource": resource},
	)
}
