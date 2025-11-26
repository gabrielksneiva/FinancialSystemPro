package service

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrPasswordHashFailed  = errors.New("password hash failed")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
)
