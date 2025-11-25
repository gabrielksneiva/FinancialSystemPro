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

// --- Stubs for Ethereum error path coverage ---

type ethGatewayFailValidate struct{}

func (ethGatewayFailValidate) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return nil, fmt.Errorf("not used")
}
func (ethGatewayFailValidate) ValidateAddress(a string) bool { return false }
func (ethGatewayFailValidate) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (ethGatewayFailValidate) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return "", fmt.Errorf("should not broadcast")
}
func (ethGatewayFailValidate) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (ethGatewayFailValidate) ChainType() entities.BlockchainType { return entities.BlockchainEthereum }

type ethGatewayBroadcastFail struct{}

func (ethGatewayBroadcastFail) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return nil, fmt.Errorf("not used")
}
func (ethGatewayBroadcastFail) ValidateAddress(a string) bool { return true }
func (ethGatewayBroadcastFail) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (ethGatewayBroadcastFail) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return "", fmt.Errorf("broadcast fail")
}
func (ethGatewayBroadcastFail) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusPending}, nil
}
func (ethGatewayBroadcastFail) ChainType() entities.BlockchainType {
	return entities.BlockchainEthereum
}

type ethGatewayOK struct{}

func (ethGatewayOK) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: entities.BlockchainEthereum, Address: "0xUser", PublicKey: "PUB"}, nil
}
func (ethGatewayOK) ValidateAddress(a string) bool { return true }
func (ethGatewayOK) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (ethGatewayOK) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("ETHHASH"), nil
}
func (ethGatewayOK) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (ethGatewayOK) ChainType() entities.BlockchainType { return entities.BlockchainEthereum }

// Repo stubs
type ethRepoNotFound struct{}

func (ethRepoNotFound) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	return nil
}
func (ethRepoNotFound) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return nil, fmt.Errorf("not found")
}
func (ethRepoNotFound) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (ethRepoNotFound) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return false, nil
}

type ethRepoOK struct{}

func (ethRepoOK) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	return nil
}
func (ethRepoOK) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return &repositories.OnChainWallet{Address: "0xUser"}, nil
}
func (ethRepoOK) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (ethRepoOK) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return true, nil
}

// DB stubs (only methods needed)
type ethDBInsertFail struct{}

func (ethDBInsertFail) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (ethDBInsertFail) Insert(v any) error                                             { return fmt.Errorf("insert fail") }
func (ethDBInsertFail) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (ethDBInsertFail) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, nil
}
func (ethDBInsertFail) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return nil
}
func (ethDBInsertFail) Balance(userID uuid.UUID) (decimal.Decimal, error) { return decimal.Zero, nil }
func (ethDBInsertFail) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type ethDBLedgerFail struct{}

func (ethDBLedgerFail) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (ethDBLedgerFail) Insert(v any) error                                             { return nil }
func (ethDBLedgerFail) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (ethDBLedgerFail) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, nil
}
func (ethDBLedgerFail) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error {
	return fmt.Errorf("ledger fail")
}
func (ethDBLedgerFail) Balance(userID uuid.UUID) (decimal.Decimal, error) { return decimal.Zero, nil }
func (ethDBLedgerFail) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type ethDBOK struct{}

func (ethDBOK) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{ID: uuid.New()}, nil
}
func (ethDBOK) Insert(v any) error                                                    { return nil }
func (ethDBOK) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error        { return nil }
func (ethDBOK) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error)      { return nil, nil }
func (ethDBOK) Transaction(userID uuid.UUID, amount decimal.Decimal, tx string) error { return nil }
func (ethDBOK) Balance(userID uuid.UUID) (decimal.Decimal, error)                     { return decimal.Zero, nil }
func (ethDBOK) UpdateTransaction(id uuid.UUID, updates map[string]interface{}) error  { return nil }

// --- Tests ---

func TestEthereumWithdraw_WalletNotFound(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayOK{})
	svc := &TransactionService{DB: ethDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoNotFound{})
	os.Setenv("ETH_VAULT_ADDRESS", "0xVault")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "Ethereum wallet não encontrada") {
		t.Fatalf("expected wallet not found branch, got %v", err)
	}
}

func TestEthereumWithdraw_VaultNotConfigured(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayOK{})
	svc := &TransactionService{DB: ethDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoOK{})
	os.Setenv("ETH_VAULT_ADDRESS", "")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "Ethereum vault não configurado") {
		t.Fatalf("expected vault not configured error, got %v", err)
	}
}

func TestEthereumWithdraw_InvalidAddress(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayFailValidate{})
	svc := &TransactionService{DB: ethDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoOK{})
	os.Setenv("ETH_VAULT_ADDRESS", "0xVault")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "endereço inválido") {
		t.Fatalf("expected invalid address error, got %v", err)
	}
}

func TestEthereumWithdraw_Overflow(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayOK{})
	svc := &TransactionService{DB: ethDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoOK{})
	os.Setenv("ETH_VAULT_ADDRESS", "0xVault")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "priv")
	// amount causing wei > 2^63-1 (e.g., 100 ETH)
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(100), "")
	if err == nil || !strings.Contains(err.Error(), "valor muito grande") {
		t.Fatalf("expected overflow error, got %v", err)
	}
}

func TestEthereumWithdraw_InsertFail(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayOK{})
	svc := &TransactionService{DB: ethDBInsertFail{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoOK{})
	os.Setenv("ETH_VAULT_ADDRESS", "0xVault")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(2), "")
	if err == nil || !strings.Contains(err.Error(), "insert fail") {
		t.Fatalf("expected insert fail error, got %v", err)
	}
}

func TestEthereumWithdraw_LedgerFail(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayOK{})
	svc := &TransactionService{DB: ethDBLedgerFail{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoOK{})
	os.Setenv("ETH_VAULT_ADDRESS", "0xVault")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(3), "")
	if err == nil || !strings.Contains(err.Error(), "ledger fail") {
		t.Fatalf("expected ledger fail error, got %v", err)
	}
}

func TestEthereumWithdraw_BroadcastFail(t *testing.T) {
	reg := NewBlockchainRegistry(ethGatewayBroadcastFail{})
	svc := &TransactionService{DB: ethDBOK{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(ethRepoOK{})
	os.Setenv("ETH_VAULT_ADDRESS", "0xVault")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "priv")
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(4), "")
	if err == nil || !strings.Contains(err.Error(), "falha broadcast Ethereum") {
		t.Fatalf("expected broadcast fail error, got %v", err)
	}
}
