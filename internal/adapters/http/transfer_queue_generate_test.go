package http

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	r "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type mockTransactionService struct {
	transferResp *services.ServiceResponse
	transferErr  error
}

func (m mockTransactionService) Deposit(string, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return nil, nil
}
func (m mockTransactionService) Withdraw(string, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return nil, nil
}
func (m mockTransactionService) WithdrawOnChain(string, entities.BlockchainType, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return nil, nil
}
func (m mockTransactionService) WithdrawTron(string, decimal.Decimal, string) (*services.ServiceResponse, error) {
	return nil, nil
}
func (m mockTransactionService) Transfer(userID string, amount decimal.Decimal, to string, cb string) (*services.ServiceResponse, error) {
	if m.transferErr != nil {
		return nil, m.transferErr
	}
	if m.transferResp != nil {
		return m.transferResp, nil
	}
	return &services.ServiceResponse{StatusCode: fiber.StatusAccepted, Body: fiber.Map{"message": "transferred"}}, nil
}
func (m mockTransactionService) GetBalance(string) (decimal.Decimal, error)     { return decimal.Zero, nil }
func (m mockTransactionService) GetWalletInfo(uuid.UUID) (*r.WalletInfo, error) { return nil, nil }

type mockQueueManager struct {
	depositID  string
	depositErr error
}

func (m mockQueueManager) EnqueueDeposit(_ context.Context, userID, amount, cb string) (string, error) {
	if m.depositErr != nil {
		return "", m.depositErr
	}
	if m.depositID != "" {
		return m.depositID, nil
	}
	return "TASK123", nil
}
func (m mockQueueManager) EnqueueWithdraw(_ context.Context, _, _, _ string) (string, error) {
	return "", nil
}
func (m mockQueueManager) EnqueueTransfer(_ context.Context, _, _, _, _ string) (string, error) {
	return "", nil
}
func (m mockQueueManager) IsConnected() bool { return true }

func run(app *fiber.App, method, path, body string) (int, string) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, _ := app.Test(req, -1)
	if resp == nil {
		return 0, ""
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, string(b)
}
func has(s, sub string) bool { return strings.Contains(s, sub) }

func TestTransfer_MissingUserID(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", h.Transfer)
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"10","to":"a@b.com"}`)
	if status != fiber.StatusBadRequest || !has(body, "user_id not found") {
		t.Fatalf("expected 400 missing user_id got %d body=%s", status, body)
	}
}
func TestTransfer_InvalidUserIDType(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", 123); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"10","to":"a@b.com"}`)
	if status != fiber.StatusBadRequest || !has(body, "Invalid user ID") {
		t.Fatalf("expected 400 invalid user ID got %d body=%s", status, body)
	}
}
func TestTransfer_InvalidJSON(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", "uid123"); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{`)
	if status != fiber.StatusBadRequest || !has(body, "Invalid JSON") && !has(body, "Validation") {
		t.Fatalf("expected validation error got %d body=%s", status, body)
	}
}
func TestTransfer_InvalidAmountFormat(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", "uid123"); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"abc","to":"a@b.com"}`)
	if status != fiber.StatusBadRequest || !has(body, "Amount must be numeric") {
		t.Fatalf("expected 400 numeric validation got %d body=%s", status, body)
	}
}
func TestTransfer_AmountZero(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", "uid123"); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"0","to":"a@b.com"}`)
	if status != fiber.StatusBadRequest || !has(body, "greater than zero") {
		t.Fatalf("expected 400 amount >0 got %d body=%s", status, body)
	}
}
func TestTransfer_MissingToValidation(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", "uid123"); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"10"}`)
	if status != fiber.StatusBadRequest || !has(body, "Validation") {
		t.Fatalf("expected 400 validation missing to got %d body=%s", status, body)
	}
}
func TestTransfer_ServiceError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{transferErr: errors.New("fail")}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", "uid123"); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"10","to":"a@b.com"}`)
	if status != fiber.StatusInternalServerError || !has(body, "fail") {
		t.Fatalf("expected 500 transfer fail got %d body=%s", status, body)
	}
}
func TestTransfer_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/transfer", func(c *fiber.Ctx) error { c.Locals("user_id", "uid123"); return h.Transfer(c) })
	status, body := run(app, fiber.MethodPost, "/api/transfer", `{"amount":"10","to":"a@b.com"}`)
	if status != fiber.StatusAccepted || !has(body, "transferred") {
		t.Fatalf("expected 202 transferred got %d body=%s", status, body)
	}
}

func TestQueueDeposit_ServiceUnavailable(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/queue/test-deposit", h.TestQueueDeposit)
	status, body := run(app, fiber.MethodPost, "/api/queue/test-deposit", `{"user_id":"uid","amount":"10"}`)
	if status != fiber.StatusServiceUnavailable || !has(body, "Queue manager not initialized") {
		t.Fatalf("expected 503 queue not initialized got %d body=%s", status, body)
	}
}
func TestQueueDeposit_InvalidBody(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, mockQueueManager{}, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/queue/test-deposit", h.TestQueueDeposit)
	status, body := run(app, fiber.MethodPost, "/api/queue/test-deposit", `{`)
	if status != fiber.StatusBadRequest || !has(body, "Invalid request body") {
		t.Fatalf("expected 400 invalid body got %d body=%s", status, body)
	}
}
func TestQueueDeposit_MissingFields(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, mockQueueManager{}, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/queue/test-deposit", h.TestQueueDeposit)
	status, body := run(app, fiber.MethodPost, "/api/queue/test-deposit", `{"user_id":""}`)
	if status != fiber.StatusBadRequest || !has(body, "user_id and amount are required") {
		t.Fatalf("expected 400 missing fields got %d body=%s", status, body)
	}
}
func TestQueueDeposit_EnqueueError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, mockQueueManager{depositErr: errors.New("qerr")}, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/queue/test-deposit", h.TestQueueDeposit)
	status, body := run(app, fiber.MethodPost, "/api/queue/test-deposit", `{"user_id":"uid","amount":"10"}`)
	if status != fiber.StatusInternalServerError || !has(body, "failed to enqueue task") {
		t.Fatalf("expected 500 enqueue error got %d body=%s", status, body)
	}
}
func TestQueueDeposit_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, mockQueueManager{depositID: "TASKX"}, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/queue/test-deposit", h.TestQueueDeposit)
	status, body := run(app, fiber.MethodPost, "/api/queue/test-deposit", `{"user_id":"uid","amount":"10"}`)
	if status != fiber.StatusAccepted || !has(body, "TASKX") {
		t.Fatalf("expected 202 TASKX got %d body=%s", status, body)
	}
}

func TestGenerateWallet_MissingUserID(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", h.GenerateWallet)
	status, body := run(app, fiber.MethodPost, "/api/v1/wallets/generate", `{"chain":"tron"}`)
	if status != fiber.StatusBadRequest || !has(body, "user_id not found") {
		t.Fatalf("expected 400 missing user_id got %d body=%s", status, body)
	}
}
func TestGenerateWallet_InvalidUUID(t *testing.T) {
	svc := &services.MultiChainWalletService{Registry: &services.BlockchainRegistry{}}
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, svc)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", func(c *fiber.Ctx) error { c.Locals("user_id", "not-a-uuid"); return h.GenerateWallet(c) })
	status, body := run(app, fiber.MethodPost, "/api/v1/wallets/generate", `{"chain":"tron"}`)
	if status != fiber.StatusBadRequest || !has(body, "Invalid user UUID") {
		t.Fatalf("expected 400 invalid uuid got %d body=%s", status, body)
	}
}
func TestGenerateWallet_ServiceUnavailable(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, nil)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.GenerateWallet(c) })
	status, body := run(app, fiber.MethodPost, "/api/v1/wallets/generate", `{"chain":"tron"}`)
	if status != fiber.StatusServiceUnavailable || !has(body, "Multi-chain wallet service not available") {
		t.Fatalf("expected 503 svc unavailable got %d body=%s", status, body)
	}
}
func TestGenerateWallet_ValidationErrorEmptyChain(t *testing.T) {
	svc := &services.MultiChainWalletService{Registry: &services.BlockchainRegistry{}}
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, svc)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.GenerateWallet(c) })
	status, body := run(app, fiber.MethodPost, "/api/v1/wallets/generate", `{}`)
	if status != fiber.StatusBadRequest || !(has(body, "VALIDATION_ERROR") || has(body, "Chain validation failed")) {
		t.Fatalf("expected 400 chain validation error got %d body=%s", status, body)
	}
}
func TestGenerateWallet_UnsupportedChain(t *testing.T) {
	svc := &services.MultiChainWalletService{Registry: &services.BlockchainRegistry{}}
	h := NewHandlerForTesting(nil, nil, mockTransactionService{}, nil, nil, newMockLogger(), nil, svc)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", func(c *fiber.Ctx) error { c.Locals("user_id", uuid.New().String()); return h.GenerateWallet(c) })
	status, body := run(app, fiber.MethodPost, "/api/v1/wallets/generate", `{"chain":"doge"}`)
	if status != fiber.StatusBadRequest || !(has(body, "VALIDATION_ERROR") || has(body, "Chain validation failed")) {
		t.Fatalf("expected 400 unsupported chain validation got %d body=%s", status, body)
	}
}
