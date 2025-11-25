package http

import (
	"net/http/httptest"
	"strings"
	"testing"

	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	domainErrors "financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"financial-system-pro/internal/shared/breaker"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Mocks for DDD transaction handler
type dddMockTxnService struct {
	depositErr  error
	withdrawErr error
	balance     decimal.Decimal
	balanceErr  error
	walletErr   error
	walletInfo  *repositories.WalletInfo
}

func (m dddMockTxnService) Deposit(userID string, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	if m.depositErr != nil {
		return nil, m.depositErr
	}
	return &services.ServiceResponse{StatusCode: fiber.StatusOK, Body: fiber.Map{"ok": true}}, nil
}
func (m dddMockTxnService) Withdraw(userID string, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	if m.withdrawErr != nil {
		return nil, m.withdrawErr
	}
	return &services.ServiceResponse{StatusCode: fiber.StatusOK, Body: fiber.Map{"ok": true}}, nil
}
func (m dddMockTxnService) WithdrawTron(userID string, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	return nil, m.withdrawErr
}
func (m dddMockTxnService) WithdrawOnChain(userID string, chain entities.BlockchainType, amount decimal.Decimal, cb string) (*services.ServiceResponse, error) {
	return nil, m.withdrawErr
}
func (m dddMockTxnService) Transfer(userID string, amount decimal.Decimal, to string, cb string) (*services.ServiceResponse, error) {
	return nil, domainErrors.NewInternalError("not implemented", nil)
}
func (m dddMockTxnService) GetBalance(userID string) (decimal.Decimal, error) {
	return m.balance, m.balanceErr
}
func (m dddMockTxnService) GetWalletInfo(id uuid.UUID) (*repositories.WalletInfo, error) {
	if m.walletErr != nil {
		return nil, m.walletErr
	}
	return m.walletInfo, nil
}

type dddMockUserService struct{}

func (dddMockUserService) CreateNewUser(req *dto.UserRequest) *domainErrors.AppError { return nil }
func (dddMockUserService) GetDatabase() services.DatabasePort                        { return nil }

// helper to create app with protected route (inject user_id)
func dddAppWith(h *DDDTransactionHandler, method, path string, fn func(*fiber.Ctx) error) *fiber.App {
	app := fiber.New()
	app.Add(method, path, func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return fn(c) })
	return app
}

func newBreakerManager() *breaker.BreakerManager { return breaker.NewBreakerManager(zap.NewNop()) }

func TestDDDTransactionHandler_Deposit_Success(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodPost, "/api/deposit", h.Deposit)
	req := httptest.NewRequest(fiber.MethodPost, "/api/deposit", strings.NewReader(`{"amount":"10.00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_Deposit_InvalidAmount(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodPost, "/api/deposit", h.Deposit)
	req := httptest.NewRequest(fiber.MethodPost, "/api/deposit", strings.NewReader(`{"amount":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_Deposit_ZeroAmount(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodPost, "/api/deposit", h.Deposit)
	req := httptest.NewRequest(fiber.MethodPost, "/api/deposit", strings.NewReader(`{"amount":"0"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_Withdraw_Success(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodPost, "/api/withdraw", h.Withdraw)
	req := httptest.NewRequest(fiber.MethodPost, "/api/withdraw", strings.NewReader(`{"amount":"5.00"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_Withdraw_InvalidAmount(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodPost, "/api/withdraw", h.Withdraw)
	req := httptest.NewRequest(fiber.MethodPost, "/api/withdraw", strings.NewReader(`{"amount":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_Transfer_Invalid(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodPost, "/api/transfer", h.Transfer)
	req := httptest.NewRequest(fiber.MethodPost, "/api/transfer", strings.NewReader(`{"amount":"2.00","to":"other"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_Balance_Success(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{balance: decimal.NewFromInt(123)}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodGet, "/api/balance", h.Balance)
	req := httptest.NewRequest(fiber.MethodGet, "/api/balance", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_GetUserWallet_NotFound(t *testing.T) {
	bm := newBreakerManager()
	h := NewDDDTransactionHandler(dddMockTxnService{walletErr: domainErrors.NewDatabaseError("not found", nil)}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodGet, "/api/wallet", h.GetUserWallet)
	req := httptest.NewRequest(fiber.MethodGet, "/api/wallet", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", resp.StatusCode)
	}
}

func TestDDDTransactionHandler_GetUserWallet_Success(t *testing.T) {
	bm := newBreakerManager()
	w := &repositories.WalletInfo{TronAddress: "TRON123"}
	h := NewDDDTransactionHandler(dddMockTxnService{walletInfo: w}, dddMockUserService{}, zap.NewNop(), bm)
	app := dddAppWith(h, fiber.MethodGet, "/api/wallet", h.GetUserWallet)
	req := httptest.NewRequest(fiber.MethodGet, "/api/wallet", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}
