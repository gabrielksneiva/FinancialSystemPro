package services

import (
	"context"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	r "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Interfaces para permitir dependency injection e mocking

// UserServiceInterface define os métodos do serviço de usuários
type UserServiceInterface interface {
	CreateNewUser(userRequest *dto.UserRequest) *errors.AppError
	GetDatabase() *r.NewDatabase
}

// AuthServiceInterface define os métodos do serviço de autenticação
type AuthServiceInterface interface {
	Login(loginData *dto.LoginRequest) (string, *errors.AppError)
}

// TransactionServiceInterface define os métodos do serviço de transações
type TransactionServiceInterface interface {
	Deposit(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error)
	Withdraw(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error)
	WithdrawTron(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error)
	Transfer(c *fiber.Ctx, amount decimal.Decimal, to string, callbackURL string) (*ServiceResponse, error)
	GetBalance(c *fiber.Ctx, userID string) (decimal.Decimal, error)
	GetWalletInfo(userID uuid.UUID) (*r.WalletInfo, error)
}

// TronServiceInterface define os métodos do serviço Tron
type TronServiceInterface interface {
	GetBalance(address string) (int64, error)
	SendTransaction(fromAddress, toAddress string, amount int64, privateKey string) (string, error)
	GetTransactionStatus(txHash string) (string, error)
	CreateWallet() (*entities.TronWallet, error)
	IsTestnetConnected() bool
	GetNetworkInfo() (map[string]interface{}, error)
	EstimateGasForTransaction(fromAddress, toAddress string, amount int64) (int64, error)
	ValidateAddress(address string) bool
	GetConnectionStatus() map[string]interface{}
	GetRPCClient() *RPCClient
	RecordError(err error)
	HealthCheck(ctx context.Context) error
}

// Garantir que os serviços concretos implementam as interfaces
var _ UserServiceInterface = (*NewUserService)(nil)
var _ AuthServiceInterface = (*NewAuthService)(nil)
var _ TransactionServiceInterface = (*NewTransactionService)(nil)
var _ TronServiceInterface = (*TronService)(nil)
