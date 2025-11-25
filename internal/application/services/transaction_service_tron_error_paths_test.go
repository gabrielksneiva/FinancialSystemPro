package services

import (
	"context"
	"encoding/json"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Fake DB variants to trigger specific TRON error branches
type fakeDBWalletMissing struct{}

func (fakeDBWalletMissing) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (fakeDBWalletMissing) Insert(v any) error { return nil }
func (fakeDBWalletMissing) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error {
	return nil
}
func (fakeDBWalletMissing) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, fmt.Errorf("wallet not found")
}
func (fakeDBWalletMissing) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return nil
}
func (fakeDBWalletMissing) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.NewFromInt(100), nil
}
func (fakeDBWalletMissing) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type fakeDBLedgerFail struct{}

func (fakeDBLedgerFail) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (fakeDBLedgerFail) Insert(v any) error                                             { return nil }
func (fakeDBLedgerFail) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (fakeDBLedgerFail) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return &repositories.WalletInfo{TronAddress: "TRONADDR"}, nil
}
func (fakeDBLedgerFail) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return fmt.Errorf("apply fail")
}
func (fakeDBLedgerFail) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.NewFromInt(50), nil
}
func (fakeDBLedgerFail) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type fakeDBBroadcastOK struct{}

func (fakeDBBroadcastOK) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (fakeDBBroadcastOK) Insert(v any) error                                             { return nil }
func (fakeDBBroadcastOK) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (fakeDBBroadcastOK) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return &repositories.WalletInfo{TronAddress: "TRONADDR"}, nil
}
func (fakeDBBroadcastOK) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return nil
}
func (fakeDBBroadcastOK) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.NewFromInt(50), nil
}
func (fakeDBBroadcastOK) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

// Tron port stubs
type tronPortConfigured struct{}

func (tronPortConfigured) SendTransaction(from, to string, amount int64, pk string) (string, error) {
	return "HASH", nil
}
func (tronPortConfigured) HasVaultConfigured() bool   { return true }
func (tronPortConfigured) GetVaultAddress() string    { return "VAULT" }
func (tronPortConfigured) GetVaultPrivateKey() string { return "PRIV" }

type tronPortSendFail struct{}

func (tronPortSendFail) SendTransaction(from, to string, amount int64, pk string) (string, error) {
	return "", fmt.Errorf("broadcast fail")
}
func (tronPortSendFail) HasVaultConfigured() bool   { return true }
func (tronPortSendFail) GetVaultAddress() string    { return "VAULT" }
func (tronPortSendFail) GetVaultPrivateKey() string { return "PRIV" }

func TestWithdrawOnChain_Tron_WalletMissing(t *testing.T) {
	svc := &TransactionService{DB: fakeDBWalletMissing{}, Tron: tronPortConfigured{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "TRON wallet not found") {
		t.Fatalf("expected wallet not found error, got %v", err)
	}
}

func TestWithdrawOnChain_Tron_LedgerApplyFail(t *testing.T) {
	svc := &TransactionService{DB: fakeDBLedgerFail{}, Tron: tronPortConfigured{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(2), "")
	if err == nil || !strings.Contains(err.Error(), "apply fail") {
		t.Fatalf("expected apply fail error, got %v", err)
	}
}

func TestWithdrawOnChain_Tron_BroadcastFail(t *testing.T) {
	svc := &TransactionService{DB: fakeDBBroadcastOK{}, Tron: tronPortSendFail{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(3), "")
	if err == nil || !strings.Contains(err.Error(), "failed to broadcast") {
		t.Fatalf("expected broadcast fail error, got %v", err)
	}
}

func TestWithdrawOnChain_Tron_Success_WithWebhook(t *testing.T) {
	var statuses []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if s, ok := payload["status"].(string); ok {
			statuses = append(statuses, s)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	svc := &TransactionService{DB: fakeDBBroadcastOK{}, Tron: tronPortConfigured{}}
	resp, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(5), server.URL)
	if err != nil || resp.StatusCode != 202 {
		t.Fatalf("expected success 202, got resp=%v err=%v", resp, err)
	}
	if len(statuses) < 3 {
		t.Fatalf("expected >=3 webhook calls, got %d (%v)", len(statuses), statuses)
	}
	hasPending := false
	hasBroadcasting := false
	hasSuccess := false
	for _, s := range statuses {
		if s == "pending" {
			hasPending = true
		}
		if s == "broadcasting" {
			hasBroadcasting = true
		}
		if s == "broadcast_success" {
			hasSuccess = true
		}
	}
	if !(hasPending && hasBroadcasting && hasSuccess) {
		t.Fatalf("missing expected statuses in %v", statuses)
	}
}

func TestWithdrawOnChain_Tron_LedgerFail_WithWebhook(t *testing.T) {
	var statuses []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if s, ok := payload["status"].(string); ok {
			statuses = append(statuses, s)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	svc := &TransactionService{DB: fakeDBLedgerFail{}, Tron: tronPortConfigured{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(2), server.URL)
	if err == nil || !strings.Contains(err.Error(), "apply fail") {
		t.Fatalf("expected apply fail error, got %v", err)
	}
	if len(statuses) < 2 {
		t.Fatalf("expected >=2 webhook calls, got %d", len(statuses))
	}
	if statuses[len(statuses)-1] != "failed" {
		t.Fatalf("expected last status failed, got %v", statuses)
	}
}

func TestWithdrawOnChain_Tron_BroadcastFail_WithWebhook(t *testing.T) {
	var statuses []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if s, ok := payload["status"].(string); ok {
			statuses = append(statuses, s)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	svc := &TransactionService{DB: fakeDBBroadcastOK{}, Tron: tronPortSendFail{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(3), server.URL)
	if err == nil || !strings.Contains(err.Error(), "failed to broadcast") {
		t.Fatalf("expected broadcast fail error, got %v", err)
	}
	if len(statuses) < 3 {
		t.Fatalf("expected >=3 webhook calls, got %d", len(statuses))
	}
	if statuses[len(statuses)-1] != "failed" {
		t.Fatalf("expected last status failed, got %v", statuses)
	}
}

// Gateway stub to exercise registry broadcast path
type tronGatewayStub struct{}

func (tronGatewayStub) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: entities.BlockchainTRON, Address: "ADDR", PublicKey: "PUB"}, nil
}
func (tronGatewayStub) ValidateAddress(a string) bool { return true }
func (tronGatewayStub) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt, EstimatedFee: 1, FeeAsset: "TRX", Source: "stub"}, nil
}
func (tronGatewayStub) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("GW_HASH"), nil
}
func (tronGatewayStub) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (tronGatewayStub) ChainType() entities.BlockchainType { return entities.BlockchainTRON }

func TestWithdrawOnChain_Tron_GatewayBroadcastSuccess(t *testing.T) {
	reg := NewBlockchainRegistry(tronGatewayStub{})
	svc := &TransactionService{DB: fakeDBBroadcastOK{}, Tron: tronPortSendFail{}} // tron port fails forcing registry path
	svc.WithChainRegistry(reg)
	resp, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(4), "")
	if err != nil || resp.StatusCode != 202 {
		t.Fatalf("expected 202 via gateway broadcast, got resp=%v err=%v", resp, err)
	}
	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map body")
	}
	if body["tx_hash"] != "GW_HASH" {
		t.Fatalf("expected tx_hash GW_HASH, got %v", body["tx_hash"])
	}
	if body["status"] != "broadcast_success" {
		t.Fatalf("expected broadcast_success status, got %v", body["status"])
	}
}
