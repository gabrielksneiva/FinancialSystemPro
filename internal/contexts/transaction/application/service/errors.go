package service

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrUserNotFound        = errors.New("user not found")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrCircuitBreakerOpen  = errors.New("circuit breaker open - service temporarily unavailable")
)
