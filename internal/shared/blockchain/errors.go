package blockchain

import "fmt"

// Error types para blockchain operations
var (
	ErrInvalidAddress        = fmt.Errorf("invalid blockchain address")
	ErrInvalidPrivateKey     = fmt.Errorf("invalid private key")
	ErrInsufficientBalance   = fmt.Errorf("insufficient balance")
	ErrTransactionFailed     = fmt.Errorf("transaction failed")
	ErrTransactionNotFound   = fmt.Errorf("transaction not found")
	ErrBlockNotFound         = fmt.Errorf("block not found")
	ErrProviderNotRegistered = fmt.Errorf("provider not registered")
	ErrProviderUnavailable   = fmt.Errorf("provider unavailable")
	ErrNetworkError          = fmt.Errorf("network error")
	ErrInvalidAmount         = fmt.Errorf("invalid amount")
	ErrInvalidNonce          = fmt.Errorf("invalid nonce")
	ErrGasEstimationFailed   = fmt.Errorf("gas estimation failed")
	ErrSignatureFailed       = fmt.Errorf("signature failed")
	ErrBroadcastFailed       = fmt.Errorf("broadcast failed")
	ErrNotSupported          = fmt.Errorf("operation not supported")
)

// ChainError representa um erro específico de blockchain
type ChainError struct {
	ChainType ChainType
	Operation string
	Err       error
}

func (e *ChainError) Error() string {
	return fmt.Sprintf("%s error on %s: %v", e.ChainType, e.Operation, e.Err)
}

func (e *ChainError) Unwrap() error {
	return e.Err
}

// NewChainError cria um novo ChainError
func NewChainError(chainType ChainType, operation string, err error) *ChainError {
	return &ChainError{
		ChainType: chainType,
		Operation: operation,
		Err:       err,
	}
}

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// NewValidationError cria um novo ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
