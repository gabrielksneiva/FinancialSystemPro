package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"

	"github.com/gofiber/fiber/v2"
)

// tronServiceMockBalance mocks TronServiceInterface methods used by GetTronBalance
type tronFullMock struct {
	validateFn      func(string) bool
	balance         int64
	balanceErr      error
	sendTxHash      string
	sendTxErr       error
	txStatus        string
	txStatusErr     error
	createWalletErr error
	connected       bool
	networkInfo     map[string]interface{}
	networkInfoErr  error
	estimateGas     int64
	estimateGasErr  error
	rpcClient       *services.RPCClient
}

func (t *tronFullMock) GetBalance(string) (int64, error) { return t.balance, t.balanceErr }
func (t *tronFullMock) SendTransaction(string, string, int64, string) (string, error) {
	return t.sendTxHash, t.sendTxErr
}
func (t *tronFullMock) GetTransactionStatus(string) (string, error) { return t.txStatus, t.txStatusErr }
func (t *tronFullMock) CreateWallet() (*entities.TronWallet, error) {
	if t.createWalletErr != nil {
		return nil, t.createWalletErr
	}
	return &entities.TronWallet{Address: "ADDR", PublicKey: "PUB"}, nil
}
func (t *tronFullMock) IsTestnetConnected() bool { return t.connected }
func (t *tronFullMock) GetNetworkInfo() (map[string]interface{}, error) {
	if t.networkInfoErr != nil {
		return nil, t.networkInfoErr
	}
	if t.networkInfo == nil {
		return map[string]interface{}{"ping": "ok"}, nil
	}
	return t.networkInfo, nil
}
func (t *tronFullMock) EstimateGasForTransaction(string, string, int64) (int64, error) {
	return t.estimateGas, t.estimateGasErr
}
func (t *tronFullMock) ValidateAddress(addr string) bool {
	if t.validateFn != nil {
		return t.validateFn(addr)
	}
	return !strings.Contains(addr, "INVALID") && addr != ""
}
func (t *tronFullMock) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{"rpc": "up"}
}
func (t *tronFullMock) GetRPCClient() *services.RPCClient { return t.rpcClient }
func (t *tronFullMock) RecordError(error)                 {}
func (t *tronFullMock) HealthCheck(context.Context) error { return nil }

// helper to execute request and return status/body
func perform(app *fiber.App, method, path string) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	resp, _ := app.Test(req, -1)
	if resp == nil {
		return 0, ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestGetTronBalance_MissingAddress(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/balance", h.GetTronBalance)

	status, body := perform(app, fiber.MethodGet, "/api/tron/balance")
	if status != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if body == "" || !contains(body, "address is required") {
		t.Fatalf("expected error message for missing address, body=%s", body)
	}
}

func TestGetTronBalance_InvalidAddress(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{validateFn: func(string) bool { return false }}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/balance", h.GetTronBalance)

	status, body := perform(app, fiber.MethodGet, "/api/tron/balance?address=TINVALID")
	if status != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
	if !contains(body, "invalid Tron address") {
		t.Fatalf("expected invalid address message, body=%s", body)
	}
}

func TestGetTronBalance_ServiceError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{balanceErr: errors.New("boom")}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/balance", h.GetTronBalance)

	status, body := perform(app, fiber.MethodGet, "/api/tron/balance?address=TVALID123")
	if status != fiber.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", status)
	}
	if !contains(body, "boom") {
		t.Fatalf("expected service error message, body=%s", body)
	}
}

func TestGetTronBalance_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{balance: 1500000}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/balance", h.GetTronBalance)

	status, body := perform(app, fiber.MethodGet, "/api/tron/balance?address=TVALID123")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if !contains(body, "balance_sun") || !contains(body, "1500000") {
		t.Fatalf("expected balance_sun 1500000 in body=%s", body)
	}
	if !contains(body, "balance_trx") || !contains(body, "1.5") {
		t.Fatalf("expected balance_trx 1.5 in body=%s", body)
	}
}

// contains helper avoids importing strings for minimal footprint
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// --- SendTronTransaction tests ---
func performWithBody(app *fiber.App, method, path, body string) (int, string) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp == nil {
		return 0, ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestSendTronTransaction_InvalidBody(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/send", h.SendTronTransaction)
	status, body := performWithBody(app, fiber.MethodPost, "/api/tron/send", "{")
	if status != fiber.StatusBadRequest || !contains(body, "Invalid JSON") {
		// dto.ValidateRequest returns Validation error message
		if !contains(body, "Validation") {
			t.Fatalf("expected validation error, got %d body=%s", status, body)
		}
	}
}

func TestSendTronTransaction_InvalidFrom(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{validateFn: func(a string) bool { return !strings.Contains(a, "BAD") }}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/send", h.SendTronTransaction)
	body := `{"from_address":"TBAD","to_address":"TGOOD","private_key":"K","amount":1}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/send", body)
	if status != fiber.StatusBadRequest || !contains(b, "invalid from address") {
		t.Fatalf("expected invalid from address 400 got %d body=%s", status, b)
	}
}

func TestSendTronTransaction_InvalidTo(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{validateFn: func(a string) bool { return !strings.Contains(a, "BAD") }}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/send", h.SendTronTransaction)
	body := `{"from_address":"TGOOD","to_address":"TBAD","private_key":"K","amount":1}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/send", body)
	if status != fiber.StatusBadRequest || !contains(b, "invalid to address") {
		t.Fatalf("expected invalid to address 400 got %d body=%s", status, b)
	}
}

func TestSendTronTransaction_ServiceError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{sendTxErr: errors.New("fail")}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/send", h.SendTronTransaction)
	body := `{"from_address":"TA","to_address":"TB","private_key":"K","amount":1}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/send", body)
	if status != fiber.StatusInternalServerError || !contains(b, "fail") {
		t.Fatalf("expected 500 fail got %d body=%s", status, b)
	}
}

func TestSendTronTransaction_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{sendTxHash: "0xHASH"}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/send", h.SendTronTransaction)
	body := `{"from_address":"TA","to_address":"TB","private_key":"K","amount":2}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/send", body)
	if status != fiber.StatusAccepted || !contains(b, "0xHASH") || !contains(b, "\"amount\":2") {
		if !contains(b, "0xHASH") {
			t.Fatalf("expected tx hash in body=%s", b)
		}
	}
}

// --- GetTronTransactionStatus tests ---
func TestGetTronTransactionStatus_MissingHash(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/tx-status", h.GetTronTransactionStatus)
	status, b := perform(app, fiber.MethodGet, "/api/tron/tx-status")
	if status != fiber.StatusBadRequest || !contains(b, "tx_hash is required") {
		t.Fatalf("expected 400 missing tx_hash got %d body=%s", status, b)
	}
}

func TestGetTronTransactionStatus_ServiceError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{txStatusErr: errors.New("oops")}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/tx-status", h.GetTronTransactionStatus)
	status, b := perform(app, fiber.MethodGet, "/api/tron/tx-status?tx_hash=ABCD")
	if status != fiber.StatusInternalServerError || !contains(b, "oops") {
		t.Fatalf("expected 500 oops got %d body=%s", status, b)
	}
}

func TestGetTronTransactionStatus_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{txStatus: "CONFIRMED"}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/tx-status", h.GetTronTransactionStatus)
	status, b := perform(app, fiber.MethodGet, "/api/tron/tx-status?tx_hash=ABCD")
	if status != fiber.StatusOK || !contains(b, "CONFIRMED") {
		t.Fatalf("expected 200 confirmed got %d body=%s", status, b)
	}
}

// --- CreateTronWallet tests ---
func TestCreateTronWallet_ServiceError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{createWalletErr: errors.New("x")}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/wallet", h.CreateTronWallet)
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/wallet", "{}")
	if status != fiber.StatusInternalServerError || !contains(b, "x") {
		t.Fatalf("expected 500 wallet error got %d body=%s", status, b)
	}
}

func TestCreateTronWallet_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/wallet", h.CreateTronWallet)
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/wallet", "{}")
	if status != fiber.StatusCreated || !contains(b, "ADDR") {
		t.Fatalf("expected 201 wallet addr got %d body=%s", status, b)
	}
}

// --- CheckTronNetwork tests ---
func TestCheckTronNetwork_Disconnected(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{connected: false}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/network", h.CheckTronNetwork)
	status, b := perform(app, fiber.MethodGet, "/api/tron/network")
	if status != fiber.StatusServiceUnavailable || !contains(b, "disconnected") {
		t.Fatalf("expected 503 disconnected got %d body=%s", status, b)
	}
}

func TestCheckTronNetwork_InfoError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{connected: true, networkInfoErr: errors.New("neterr")}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/network", h.CheckTronNetwork)
	status, b := perform(app, fiber.MethodGet, "/api/tron/network")
	if status != fiber.StatusInternalServerError || !contains(b, "neterr") {
		t.Fatalf("expected 500 neterr got %d body=%s", status, b)
	}
}

func TestCheckTronNetwork_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{connected: true, networkInfo: map[string]interface{}{"height": 10}}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/network", h.CheckTronNetwork)
	status, b := perform(app, fiber.MethodGet, "/api/tron/network")
	if status != fiber.StatusOK || !contains(b, "height") {
		t.Fatalf("expected 200 network info got %d body=%s", status, b)
	}
}

// --- EstimateTronGas tests ---
func TestEstimateTronGas_InvalidBody(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/estimate-energy", h.EstimateTronGas)
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/estimate-energy", "{")
	if status != fiber.StatusBadRequest || !contains(b, "Invalid request body") {
		t.Fatalf("expected 400 invalid body got %d body=%s", status, b)
	}
}

func TestEstimateTronGas_InvalidFrom(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{validateFn: func(a string) bool { return !strings.Contains(a, "BAD") }}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/estimate-energy", h.EstimateTronGas)
	body := `{"from_address":"TBAD","to_address":"TGOOD","amount":1}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/estimate-energy", body)
	if status != fiber.StatusBadRequest || !contains(b, "invalid from address") {
		t.Fatalf("expected 400 invalid from got %d body=%s", status, b)
	}
}

func TestEstimateTronGas_InvalidTo(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{validateFn: func(a string) bool { return !strings.Contains(a, "BAD") }}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/estimate-energy", h.EstimateTronGas)
	body := `{"from_address":"TGOOD","to_address":"TBAD","amount":1}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/estimate-energy", body)
	if status != fiber.StatusBadRequest || !contains(b, "invalid to address") {
		t.Fatalf("expected 400 invalid to got %d body=%s", status, b)
	}
}

func TestEstimateTronGas_ServiceError(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{estimateGasErr: errors.New("egerr")}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/estimate-energy", h.EstimateTronGas)
	body := `{"from_address":"TA","to_address":"TB","amount":2}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/estimate-energy", body)
	if status != fiber.StatusInternalServerError || !contains(b, "egerr") {
		t.Fatalf("expected 500 egerr got %d body=%s", status, b)
	}
}

func TestEstimateTronGas_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{estimateGas: 999}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/estimate-energy", h.EstimateTronGas)
	body := `{"from_address":"TA","to_address":"TB","amount":3}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/estimate-energy", body)
	if status != fiber.StatusOK || !contains(b, "999") {
		t.Fatalf("expected 200 energy 999 got %d body=%s", status, b)
	}
}

// --- GetRPCStatus & GetAvailableMethods tests ---
func TestGetRPCStatus_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/rpc-status", h.GetRPCStatus)
	status, b := perform(app, fiber.MethodGet, "/api/tron/rpc-status")
	if status != fiber.StatusOK || !contains(b, "rpc") {
		t.Fatalf("expected 200 rpc status got %d body=%s", status, b)
	}
}

func TestGetAvailableMethods_Success(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Get("/api/tron/rpc-methods", h.GetAvailableMethods)
	status, b := perform(app, fiber.MethodGet, "/api/tron/rpc-methods")
	if status != fiber.StatusOK || !contains(b, "available_methods") {
		t.Fatalf("expected 200 methods got %d body=%s", status, b)
	}
}

// --- CallRPCMethod tests ---
func TestCallRPCMethod_NoClient(t *testing.T) {
	h := NewHandlerForTesting(nil, nil, &tronFullMock{rpcClient: nil}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/rpc-call", h.CallRPCMethod)
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/rpc-call", `{"method":"eth_blockNumber"}`)
	if status != fiber.StatusServiceUnavailable || !contains(b, "not available") {
		t.Fatalf("expected 503 no client got %d body=%s", status, b)
	}
}

func TestCallRPCMethod_CallError(t *testing.T) {
	// server returns 500 to trigger error path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("boom"))
	}))
	defer server.Close()
	rpc := services.NewRPCClient(server.URL)
	h := NewHandlerForTesting(nil, nil, &tronFullMock{rpcClient: rpc}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/rpc-call", h.CallRPCMethod)
	body := `{"method":"eth_blockNumber"}`
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/rpc-call", body)
	if status != fiber.StatusInternalServerError || !contains(b, "status HTTP") {
		t.Fatalf("expected 500 rpc error got %d body=%s", status, b)
	}
}

func TestCallRPCMethod_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		resp := map[string]interface{}{"jsonrpc": "2.0", "result": json.RawMessage(`"ok"`), "id": time.Now().Unix()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	rpc := services.NewRPCClient(server.URL)
	h := NewHandlerForTesting(nil, nil, &tronFullMock{rpcClient: rpc}, nil, newMockLogger(), nil)
	app := fiber.New()
	app.Post("/api/tron/rpc-call", h.CallRPCMethod)
	status, b := performWithBody(app, fiber.MethodPost, "/api/tron/rpc-call", `{"method":"eth_blockNumber"}`)
	if status != fiber.StatusOK || !contains(b, "ok") {
		t.Fatalf("expected 200 ok got %d body=%s", status, b)
	}
}
