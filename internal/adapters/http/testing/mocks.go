package testing

import (
	"context"
	"financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repo "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MockUserService implementa services.UserServiceInterface para testes
type MockUserService struct {
	CreateNewUserFunc func(*dto.UserRequest) *errors.AppError
	GetDatabaseFunc   func() services.DatabasePort
}

func (m *MockUserService) CreateNewUser(req *dto.UserRequest) *errors.AppError {
	if m.CreateNewUserFunc != nil {
		return m.CreateNewUserFunc(req)
	}
	return nil
}

func (m *MockUserService) GetDatabase() services.DatabasePort {
	if m.GetDatabaseFunc != nil {
		return m.GetDatabaseFunc()
	}
	return nil
}

// MockAuthService implementa services.AuthServiceInterface para testes
type MockAuthService struct {
	LoginFunc func(*dto.LoginRequest) (string, *errors.AppError)
}

func (m *MockAuthService) Login(loginData *dto.LoginRequest) (string, *errors.AppError) {
	if m.LoginFunc != nil {
		return m.LoginFunc(loginData)
	}
	return "mock-token-123", nil
}

// MockTransactionService implementa services.TransactionServiceInterface para testes
type MockTransactionService struct {
	DepositFunc         func(string, decimal.Decimal, string) (*services.ServiceResponse, error)
	WithdrawFunc        func(string, decimal.Decimal, string) (*services.ServiceResponse, error)
	WithdrawOnChainFunc func(string, entities.BlockchainType, decimal.Decimal, string) (*services.ServiceResponse, error)
	WithdrawTronFunc    func(string, decimal.Decimal, string) (*services.ServiceResponse, error)
	TransferFunc        func(string, decimal.Decimal, string, string) (*services.ServiceResponse, error)
	GetBalanceFunc      func(string) (decimal.Decimal, error)
	GetWalletInfoFunc   func(uuid.UUID) (*repo.WalletInfo, error)
}

func (m *MockTransactionService) Deposit(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
	if m.DepositFunc != nil {
		return m.DepositFunc(userID, amount, callbackURL)
	}
	return &services.ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"success": true}}, nil
}

func (m *MockTransactionService) Withdraw(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(userID, amount, callbackURL)
	}
	return &services.ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"success": true}}, nil
}

func (m *MockTransactionService) WithdrawTron(userID string, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
	if m.WithdrawTronFunc != nil {
		return m.WithdrawTronFunc(userID, amount, callbackURL)
	}
	return &services.ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"success": true}}, nil
}

func (m *MockTransactionService) WithdrawOnChain(userID string, chain entities.BlockchainType, amount decimal.Decimal, callbackURL string) (*services.ServiceResponse, error) {
	if m.WithdrawOnChainFunc != nil {
		return m.WithdrawOnChainFunc(userID, chain, amount, callbackURL)
	}
	return &services.ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"success": true, "chain": chain}}, nil
}

func (m *MockTransactionService) Transfer(userID string, amount decimal.Decimal, to string, callbackURL string) (*services.ServiceResponse, error) {
	if m.TransferFunc != nil {
		return m.TransferFunc(userID, amount, to, callbackURL)
	}
	return &services.ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"success": true}}, nil
}

func (m *MockTransactionService) GetBalance(userID string) (decimal.Decimal, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(userID)
	}
	return decimal.NewFromInt(1000), nil
}

func (m *MockTransactionService) GetWalletInfo(userID uuid.UUID) (*repo.WalletInfo, error) {
	if m.GetWalletInfoFunc != nil {
		return m.GetWalletInfoFunc(userID)
	}
	return &repo.WalletInfo{
		UserID:      userID,
		TronAddress: "TMockAddress123456789012345678901",
	}, nil
}

// MockTronService implementa services.TronServiceInterface para testes
type MockTronService struct {
	GetBalanceFunc                func(string) (int64, error)
	SendTransactionFunc           func(string, string, int64, string) (string, error)
	GetTransactionStatusFunc      func(string) (string, error)
	CreateWalletFunc              func() (*entities.TronWallet, error)
	IsTestnetConnectedFunc        func() bool
	GetNetworkInfoFunc            func() (map[string]interface{}, error)
	EstimateGasForTransactionFunc func(string, string, int64) (int64, error)
	ValidateAddressFunc           func(string) bool
	GetConnectionStatusFunc       func() map[string]interface{}
	GetRPCClientFunc              func() *services.RPCClient
	RecordErrorFunc               func(error)
	HealthCheckFunc               func(context.Context) error
}

func (m *MockTronService) GetBalance(address string) (int64, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(address)
	}
	return 1000000, nil
}

func (m *MockTronService) SendTransaction(from, to string, amount int64, privateKey string) (string, error) {
	if m.SendTransactionFunc != nil {
		return m.SendTransactionFunc(from, to, amount, privateKey)
	}
	return "mock-tx-hash-123", nil
}

func (m *MockTronService) GetTransactionStatus(txHash string) (string, error) {
	if m.GetTransactionStatusFunc != nil {
		return m.GetTransactionStatusFunc(txHash)
	}
	return "SUCCESS", nil
}

func (m *MockTronService) CreateWallet() (*entities.TronWallet, error) {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc()
	}
	return &entities.TronWallet{
		Address:    "TMockWallet123456789012345678901",
		PrivateKey: "mock-private-key",
		PublicKey:  "mock-public-key",
	}, nil
}

func (m *MockTronService) IsTestnetConnected() bool {
	if m.IsTestnetConnectedFunc != nil {
		return m.IsTestnetConnectedFunc()
	}
	return true
}

func (m *MockTronService) GetNetworkInfo() (map[string]interface{}, error) {
	if m.GetNetworkInfoFunc != nil {
		return m.GetNetworkInfoFunc()
	}
	return map[string]interface{}{
		"network": "Testnet",
		"status":  "connected",
	}, nil
}

func (m *MockTronService) EstimateGasForTransaction(from, to string, amount int64) (int64, error) {
	if m.EstimateGasForTransactionFunc != nil {
		return m.EstimateGasForTransactionFunc(from, to, amount)
	}
	return 10000, nil
}

func (m *MockTronService) ValidateAddress(address string) bool {
	if m.ValidateAddressFunc != nil {
		return m.ValidateAddressFunc(address)
	}
	return len(address) > 0 && address[0] == 'T'
}

func (m *MockTronService) GetConnectionStatus() map[string]interface{} {
	if m.GetConnectionStatusFunc != nil {
		return m.GetConnectionStatusFunc()
	}
	return map[string]interface{}{"status": "connected"}
}

func (m *MockTronService) GetRPCClient() *services.RPCClient {
	if m.GetRPCClientFunc != nil {
		return m.GetRPCClientFunc()
	}
	return nil
}

func (m *MockTronService) RecordError(err error) {
	if m.RecordErrorFunc != nil {
		m.RecordErrorFunc(err)
	}
}

func (m *MockTronService) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

// MockQueueManager simula o gerenciador de filas
type MockQueueManager struct {
	EnqueueDepositFunc  func(userID, amount, callbackURL string) (string, error)
	EnqueueWithdrawFunc func(userID, amount, callbackURL string) (string, error)
	EnqueueTransferFunc func(fromUserID, toUserID, amount, callbackURL string) (string, error)
	IsConnectedFunc     func() bool
}

func (m *MockQueueManager) EnqueueDeposit(ctx context.Context, userID, amount, callbackURL string) (string, error) {
	if m.EnqueueDepositFunc != nil {
		return m.EnqueueDepositFunc(userID, amount, callbackURL)
	}
	return "mock-task-id", nil
}

func (m *MockQueueManager) EnqueueWithdraw(ctx context.Context, userID, amount, callbackURL string) (string, error) {
	if m.EnqueueWithdrawFunc != nil {
		return m.EnqueueWithdrawFunc(userID, amount, callbackURL)
	}
	return "mock-task-id", nil
}

func (m *MockQueueManager) EnqueueTransfer(ctx context.Context, fromUserID, toUserID, amount, callbackURL string) (string, error) {
	if m.EnqueueTransferFunc != nil {
		return m.EnqueueTransferFunc(fromUserID, toUserID, amount, callbackURL)
	}
	return "mock-task-id", nil
}

func (m *MockQueueManager) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return true
}

// MockLogger simula o logger
type MockLogger struct {
	InfoFunc  func(msg string, fields ...zap.Field)
	WarnFunc  func(msg string, fields ...zap.Field)
	ErrorFunc func(msg string, fields ...zap.Field)
	DebugFunc func(msg string, fields ...zap.Field)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	if m.InfoFunc != nil {
		m.InfoFunc(msg, fields...)
	}
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	if m.WarnFunc != nil {
		m.WarnFunc(msg, fields...)
	}
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(msg, fields...)
	}
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	if m.DebugFunc != nil {
		m.DebugFunc(msg, fields...)
	}
}

// MockRateLimiter simula o rate limiter
type MockRateLimiter struct {
	MiddlewareFunc func(action string) fiber.Handler
	IsAllowedFunc  func(userID string, action string) bool
}

func (m *MockRateLimiter) Middleware(action string) fiber.Handler {
	if m.MiddlewareFunc != nil {
		return m.MiddlewareFunc(action)
	}
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func (m *MockRateLimiter) IsAllowed(userID string, action string) bool {
	if m.IsAllowedFunc != nil {
		return m.IsAllowedFunc(userID, action)
	}
	return true
}

// SetupTestHandler cria um handler de teste com serviços mockados
func SetupTestHandler() *http.Handler {
	userMock := &MockUserService{}
	authMock := &MockAuthService{}
	txMock := &MockTransactionService{}
	tronMock := &MockTronService{}
	loggerMock := &MockLogger{}
	rateLimiterMock := &MockRateLimiter{}

	return http.NewHandlerForTesting(
		userMock,
		authMock,
		txMock,
		tronMock,
		nil,
		loggerMock,
		rateLimiterMock,
		nil,
	)
}
