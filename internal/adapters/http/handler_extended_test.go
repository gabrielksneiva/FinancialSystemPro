package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	r "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Mocks específicos adicionais
type walletTxnService struct {
	wallet  *r.WalletInfo
	balance decimal.Decimal
	wErr    error
}

func (w walletTxnService) Deposit(string, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return nil, errors.New("not impl")
}
func (w walletTxnService) Withdraw(string, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return &services.ServiceResponse{StatusCode: fiber.StatusAccepted, Body: fiber.Map{"message": "withdrawn"}}, nil
}
func (w walletTxnService) WithdrawOnChain(string, entities.BlockchainType, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return &services.ServiceResponse{StatusCode: fiber.StatusAccepted, Body: fiber.Map{"message": "onchain"}}, nil
}
func (w walletTxnService) WithdrawTron(string, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return &services.ServiceResponse{StatusCode: fiber.StatusAccepted, Body: fiber.Map{"message": "tron"}}, nil
}
func (w walletTxnService) Transfer(string, decimal.Decimal, string, string) (*services.ServiceResponse, error) {
	return nil, errors.New("not impl")
}
func (w walletTxnService) GetBalance(string) (decimal.Decimal, error)        { return w.balance, nil }
func (w walletTxnService) GetWalletInfo(id uuid.UUID) (*r.WalletInfo, error) { return w.wallet, w.wErr }

type tronServiceMock struct {
	valid   bool
	balance int64
}

func (t tronServiceMock) GetBalance(addr string) (int64, error) { return t.balance, nil }
func (t tronServiceMock) SendTransaction(f, to string, amt int64, pk string) (string, error) {
	return "hash123", nil
}
func (t tronServiceMock) GetTransactionStatus(h string) (string, error) { return "CONFIRMED", nil }
func (t tronServiceMock) CreateWallet() (*entities.TronWallet, error) {
	return &entities.TronWallet{Address: "ADDR", PrivateKey: "PK", PublicKey: "PUB"}, nil
}
func (t tronServiceMock) IsTestnetConnected() bool { return true }
func (t tronServiceMock) GetNetworkInfo() (map[string]interface{}, error) {
	return map[string]interface{}{"net": "test"}, nil
}
func (t tronServiceMock) EstimateGasForTransaction(f, to string, amt int64) (int64, error) {
	return 1, nil
}
func (t tronServiceMock) ValidateAddress(address string) bool { return t.valid }
func (t tronServiceMock) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{"rpc": true}
}
func (t tronServiceMock) GetRPCClient() *services.RPCClient     { return nil }
func (t tronServiceMock) RecordError(err error)                 {}
func (t tronServiceMock) HealthCheck(ctx context.Context) error { return nil }

type dummyRateLimiter struct{}

func (d dummyRateLimiter) Middleware(action string) fiber.Handler {
	return func(c *fiber.Ctx) error { return c.Next() }
}
func (d dummyRateLimiter) IsAllowed(string, string) bool { return true }

// Not needed: user service unused in these handler paths

// Tests
func TestHandler_Balance_Success(t *testing.T) {
	tx := walletTxnService{balance: decimal.NewFromFloat(123.45)}
	h := NewHandlerForTesting(nil, nil, tx, tronServiceMock{valid: true, balance: 1000}, nil, NewZapLoggerAdapter(zap.NewNop()), dummyRateLimiter{}, nil)
	app := fiber.New()
	app.Get("/api/balance", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.Balance(c) })
	req := httptest.NewRequest(fiber.MethodGet, "/api/balance", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandler_GetUserWallet_NotFound(t *testing.T) {
	tx := walletTxnService{wErr: errors.New("missing")}
	h := NewHandlerForTesting(nil, nil, tx, tronServiceMock{valid: true}, nil, NewZapLoggerAdapter(zap.NewNop()), dummyRateLimiter{}, nil)
	app := fiber.New()
	app.Get("/api/wallet", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.GetUserWallet(c) })
	req := httptest.NewRequest(fiber.MethodGet, "/api/wallet", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 got %d", resp.StatusCode)
	}
}

func TestHandler_GetUserWallet_Success(t *testing.T) {
	uid := uuid.New()
	tx := walletTxnService{wallet: &r.WalletInfo{UserID: uid, TronAddress: "TADDR"}}
	h := NewHandlerForTesting(nil, nil, tx, tronServiceMock{valid: true}, nil, NewZapLoggerAdapter(zap.NewNop()), dummyRateLimiter{}, nil)
	app := fiber.New()
	app.Get("/api/wallet", func(c *fiber.Ctx) error { c.Locals("user_id", uid.String()); return h.GetUserWallet(c) })
	req := httptest.NewRequest(fiber.MethodGet, "/api/wallet", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestHandler_Withdraw_InvalidAmount(t *testing.T) {
	tx := walletTxnService{}
	h := NewHandlerForTesting(nil, nil, tx, tronServiceMock{valid: true}, nil, NewZapLoggerAdapter(zap.NewNop()), dummyRateLimiter{}, nil)
	app := fiber.New()
	app.Post("/api/withdraw", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.Withdraw(c) })
	req := httptest.NewRequest(fiber.MethodPost, "/api/withdraw", strings.NewReader(`{"amount":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestHandler_Withdraw_NegativeAmount(t *testing.T) {
	tx := walletTxnService{}
	h := NewHandlerForTesting(nil, nil, tx, tronServiceMock{valid: true}, nil, NewZapLoggerAdapter(zap.NewNop()), dummyRateLimiter{}, nil)
	app := fiber.New()
	app.Post("/api/withdraw", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.Withdraw(c) })
	req := httptest.NewRequest(fiber.MethodPost, "/api/withdraw", strings.NewReader(`{"amount":"-5"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestHandler_Withdraw_InternalSuccess(t *testing.T) {
	tx := walletTxnService{}
	h := NewHandlerForTesting(nil, nil, tx, tronServiceMock{valid: true}, nil, NewZapLoggerAdapter(zap.NewNop()), dummyRateLimiter{}, nil)
	app := fiber.New()
	app.Post("/api/withdraw", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		c.Request().SetBody([]byte(`{"amount":"5"}`))
		c.Request().Header.Set("Content-Type", "application/json")
		return h.Withdraw(c)
	})
	req := httptest.NewRequest(fiber.MethodPost, "/api/withdraw", strings.NewReader(`{"amount":"5"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("expected 202 got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["message"] == nil {
		t.Fatalf("missing message")
	}
}
