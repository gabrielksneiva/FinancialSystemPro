package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	domainErrors "financial-system-pro/internal/domain/errors"
	r "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Minimal mocks implementing required interfaces for handler tests.

type mockUserService struct{ createErr *domainErrors.AppError }

func (m mockUserService) CreateNewUser(req *dto.UserRequest) *domainErrors.AppError {
	return m.createErr
}
func (m mockUserService) GetDatabase() services.DatabasePort { return nil }

type mockAuthService struct {
	token string
	err   *domainErrors.AppError
}

func (m mockAuthService) Login(req *dto.LoginRequest) (string, *domainErrors.AppError) {
	return m.token, m.err
}

type mockTxnService struct {
	depositResp   *services.ServiceResponse
	depositErr    error
	balanceResp   decimal.Decimal
	balanceErr    error
	walletInfoErr error
}

func (m mockTxnService) Deposit(userID string, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	return m.depositResp, m.depositErr
}
func (m mockTxnService) Withdraw(userID string, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	return nil, m.depositErr
}
func (m mockTxnService) WithdrawOnChain(userID string, chain entities.BlockchainType, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	return nil, m.depositErr
}
func (m mockTxnService) WithdrawTron(userID string, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	return nil, m.depositErr
}
func (m mockTxnService) Transfer(userID string, amount decimal.Decimal, to string, cb string) (*services.ServiceResponse, error) {
	return nil, m.depositErr
}
func (m mockTxnService) GetBalance(userID string) (decimal.Decimal, error) {
	return m.balanceResp, m.balanceErr
}
func (m mockTxnService) GetWalletInfo(id uuid.UUID) (*r.WalletInfo, error) {
	return nil, m.walletInfoErr
}

type noopTronService struct{}

func (noopTronService) GetBalance(a string) (int64, error) { return 0, nil }
func (noopTronService) SendTransaction(f, t string, amt int64, pk string) (string, error) {
	return "txhash", nil
}
func (noopTronService) GetTransactionStatus(h string) (string, error) { return "CONFIRMED", nil }
func (noopTronService) CreateWallet() (*entities.TronWallet, error) {
	return &entities.TronWallet{Address: "ADDR", PrivateKey: "PK", PublicKey: "PUB"}, nil
}
func (noopTronService) IsTestnetConnected() bool { return true }
func (noopTronService) GetNetworkInfo() (map[string]interface{}, error) {
	return map[string]interface{}{"net": "test"}, nil
}
func (noopTronService) EstimateGasForTransaction(f, t string, amt int64) (int64, error) {
	return 1, nil
}
func (noopTronService) ValidateAddress(address string) bool { return true }
func (noopTronService) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{"rpc": true}
}
func (noopTronService) GetRPCClient() *services.RPCClient     { return nil }
func (noopTronService) RecordError(err error)                 {}
func (noopTronService) HealthCheck(ctx context.Context) error { return nil }

type mockLogger struct{ *zap.Logger }

func newMockLogger() LoggerInterface { return NewZapLoggerAdapter(zap.NewNop()) }

type noopRateLimiter struct{}

func (noopRateLimiter) Middleware(action string) fiber.Handler {
	return func(c *fiber.Ctx) error { return c.Next() }
}
func (noopRateLimiter) IsAllowed(string, string) bool { return true }

// Helper to create fiber app with specific handler method.
func setupApp(h *Handler, method, path string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Add(method, path, handler)
	return app
}

func TestHandler_CreateUser_Success(t *testing.T) {
	h := NewHandlerForTesting(mockUserService{}, mockAuthService{}, mockTxnService{}, noopTronService{}, nil, newMockLogger(), noopRateLimiter{}, nil)
	app := setupApp(h, fiber.MethodPost, "/api/users", h.CreateUser)
	req := httptest.NewRequest(fiber.MethodPost, "/api/users", strings.NewReader(`{"email":"a@b.com","password":"StrongPass123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestHandler_CreateUser_ValidationError(t *testing.T) {
	h := NewHandlerForTesting(mockUserService{}, mockAuthService{}, mockTxnService{}, noopTronService{}, nil, newMockLogger(), noopRateLimiter{}, nil)
	app := setupApp(h, fiber.MethodPost, "/api/users", h.CreateUser)
	req := httptest.NewRequest(fiber.MethodPost, "/api/users", strings.NewReader(`{"email":"invalid","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	// Expect 400 from validator (email format)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Login_UserNotFound(t *testing.T) {
	h := NewHandlerForTesting(mockUserService{}, mockAuthService{token: "", err: domainErrors.NewValidationError("email", "Email not registered")}, mockTxnService{}, noopTronService{}, nil, newMockLogger(), noopRateLimiter{}, nil)
	app := setupApp(h, fiber.MethodPost, "/api/login", h.Login)
	req := httptest.NewRequest(fiber.MethodPost, "/api/login", strings.NewReader(`{"email":"missing@b.com","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Deposit_MissingUserID(t *testing.T) {
	h := NewHandlerForTesting(mockUserService{}, mockAuthService{}, mockTxnService{}, noopTronService{}, nil, newMockLogger(), noopRateLimiter{}, nil)
	app := setupApp(h, fiber.MethodPost, "/api/deposit", h.Deposit)
	req := httptest.NewRequest(fiber.MethodPost, "/api/deposit", strings.NewReader(`{"amount":"10.00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Deposit_InvalidAmountFormat(t *testing.T) {
	trxSvc := mockTxnService{depositResp: &services.ServiceResponse{StatusCode: fiber.StatusAccepted, Body: fiber.Map{"ok": true}}}
	h := NewHandlerForTesting(mockUserService{}, mockAuthService{}, trxSvc, noopTronService{}, nil, newMockLogger(), noopRateLimiter{}, nil)
	app := fiber.New()
	app.Post("/api/deposit", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.Deposit(c) })
	req := httptest.NewRequest(fiber.MethodPost, "/api/deposit", strings.NewReader(`{"amount":"notnumber"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Deposit_Success(t *testing.T) {
	trxSvc := mockTxnService{depositResp: &services.ServiceResponse{StatusCode: fiber.StatusAccepted, Body: fiber.Map{"message": "queued"}}}
	h := NewHandlerForTesting(mockUserService{}, mockAuthService{}, trxSvc, noopTronService{}, nil, newMockLogger(), noopRateLimiter{}, nil)
	app := fiber.New()
	app.Post("/api/deposit", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.Deposit(c) })
	req := httptest.NewRequest(fiber.MethodPost, "/api/deposit", strings.NewReader(`{"amount":"15.00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusAccepted, resp.StatusCode)
	var body map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "queued", body["message"])
}
