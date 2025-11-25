package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// --- Stubs for Bitcoin error path coverage ---

type btcGatewayFailValidate struct{}

func (btcGatewayFailValidate) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return nil, fmt.Errorf("not used")
}
func (btcGatewayFailValidate) ValidateAddress(a string) bool { return false }
func (btcGatewayFailValidate) EstimateFee(ctx context.Context, f, t string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (btcGatewayFailValidate) Broadcast(ctx context.Context, f, t string, amt int64, pk string) (TxHash, error) {
	return "", fmt.Errorf("should not broadcast")
}
func (btcGatewayFailValidate) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (btcGatewayFailValidate) ChainType() entities.BlockchainType { return entities.BlockchainBitcoin }

type btcGatewayBroadcastFail struct{}

func (btcGatewayBroadcastFail) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return nil, fmt.Errorf("not used")
}
func (btcGatewayBroadcastFail) ValidateAddress(a string) bool { return true }
func (btcGatewayBroadcastFail) EstimateFee(ctx context.Context, f, t string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (btcGatewayBroadcastFail) Broadcast(ctx context.Context, f, t string, amt int64, pk string) (TxHash, error) {
	return "", fmt.Errorf("broadcast fail")
}
func (btcGatewayBroadcastFail) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusPending}, nil
}
func (btcGatewayBroadcastFail) ChainType() entities.BlockchainType { return entities.BlockchainBitcoin }

type btcGatewayOK struct{}

func (btcGatewayOK) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: entities.BlockchainBitcoin, Address: "1UserBitcoinAddr", PublicKey: "PUB"}, nil
}
func (btcGatewayOK) ValidateAddress(a string) bool { return true }
func (btcGatewayOK) EstimateFee(ctx context.Context, f, t string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (btcGatewayOK) Broadcast(ctx context.Context, f, t string, amt int64, pk string) (TxHash, error) {
	return TxHash("BTCHASH"), nil
}
func (btcGatewayOK) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (btcGatewayOK) ChainType() entities.BlockchainType { return entities.BlockchainBitcoin }

// Repo stubs
type btcRepoNotFound struct{}

func (btcRepoNotFound) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	return nil
}
func (btcRepoNotFound) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return nil, fmt.Errorf("not found")
}
func (btcRepoNotFound) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (btcRepoNotFound) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return false, nil
}

type btcRepoOK struct{}

func (btcRepoOK) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	return nil
}
func (btcRepoOK) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return &repositories.OnChainWallet{Address: "1UserBitcoinAddr"}, nil
}
func (btcRepoOK) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (btcRepoOK) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return true, nil
}

// DB stubs
type btcDBInsertFail struct{}

func (btcDBInsertFail) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (btcDBInsertFail) Insert(v any) error                                             { return fmt.Errorf("insert fail") }
func (btcDBInsertFail) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (btcDBInsertFail) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, nil
}
func (btcDBInsertFail) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return nil
}
func (btcDBInsertFail) Balance(userID uuid.UUID) (decimal.Decimal, error) { return decimal.Zero, nil }
func (btcDBInsertFail) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type btcDBLedgerFail struct{}

func (btcDBLedgerFail) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (btcDBLedgerFail) Insert(v any) error                                             { return nil }
func (btcDBLedgerFail) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (btcDBLedgerFail) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, nil
}
func (btcDBLedgerFail) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return fmt.Errorf("ledger fail")
}
func (btcDBLedgerFail) Balance(userID uuid.UUID) (decimal.Decimal, error) { return decimal.Zero, nil }
func (btcDBLedgerFail) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type btcDBOK struct{}

func (btcDBOK) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (btcDBOK) Insert(v any) error                                                    { return nil }
func (btcDBOK) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error        { return nil }
func (btcDBOK) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error)      { return nil, nil }
func (btcDBOK) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error { return nil }
func (btcDBOK) Balance(userID uuid.UUID) (decimal.Decimal, error)                     { return decimal.Zero, nil }
func (btcDBOK) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error  { return nil }

// --- Tests ---

func TestBitcoinWithdraw_WalletNotFound(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayOK{})
	svc := &TransactionService{DB: btcDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoNotFound{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "Bitcoin wallet não encontrada") {
		t.Fatalf("expected wallet not found branch, got %v", err)
	}
}

func TestBitcoinWithdraw_VaultNotConfigured(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayOK{})
	svc := &TransactionService{DB: btcDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "Bitcoin vault não configurado") {
		t.Fatalf("expected vault not configured error, got %v", err)
	}
}

func TestBitcoinWithdraw_InvalidAddress(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayFailValidate{})
	svc := &TransactionService{DB: btcDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "endereço inválido") {
		t.Fatalf("expected invalid address error, got %v", err)
	}
}

func TestBitcoinWithdraw_Overflow(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayOK{})
	svc := &TransactionService{DB: btcDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	// amount causing satoshis > 2^63-1 (e.g., 300 BTC)
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(300), "")
	if err == nil || !strings.Contains(err.Error(), "valor muito grande") {
		t.Fatalf("expected overflow error, got %v", err)
	}
}

func TestBitcoinWithdraw_InsertFail(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayOK{})
	svc := &TransactionService{DB: btcDBInsertFail{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(2), "")
	if err == nil || !strings.Contains(err.Error(), "insert fail") {
		t.Fatalf("expected insert fail error, got %v", err)
	}
}

func TestBitcoinWithdraw_LedgerFail(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayOK{})
	svc := &TransactionService{DB: btcDBLedgerFail{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(3), "")
	if err == nil || !strings.Contains(err.Error(), "ledger fail") {
		t.Fatalf("expected ledger fail error, got %v", err)
	}
}

func TestBitcoinWithdraw_BroadcastFail(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayBroadcastFail{})
	svc := &TransactionService{DB: btcDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(4), "")
	if err == nil || !strings.Contains(err.Error(), "falha broadcast Bitcoin") {
		t.Fatalf("expected broadcast fail error, got %v", err)
	}
}

func TestBitcoinWithdraw_Success(t *testing.T) {
	reg := NewBlockchainRegistry(btcGatewayOK{})
	svc := &TransactionService{DB: btcDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(btcRepoOK{})
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultAddr")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "priv")
	resp, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainBitcoin, decimal.NewFromInt(5), "")
	if err != nil {
		t.Fatalf("unexpected error success path: %v", err)
	}
	body := resp.Body.(map[string]interface{})
	if body["status"] != "broadcast_success" || body["onchain_chain"] != "bitcoin" {
		t.Fatalf("unexpected body %+v", body)
	}
}
