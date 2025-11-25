package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// minimal fake DB implementing required methods
type fakeDBTxn struct{}

func (fakeDBTxn) FindUserByField(field string, value any) (*repositories.User, error) {
	return &repositories.User{}, nil
}
func (fakeDBTxn) Insert(v any) error                                             { return nil }
func (fakeDBTxn) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error { return nil }
func (fakeDBTxn) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return &repositories.WalletInfo{TronAddress: "ADDR"}, nil
}
func (fakeDBTxn) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return nil
}
func (fakeDBTxn) Balance(userID uuid.UUID) (decimal.Decimal, error)                      { return decimal.Zero, nil }
func (fakeDBTxn) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error { return nil }

func TestTransactionService_WithdrawOnChain_InvalidUser(t *testing.T) {
	svc := &TransactionService{DB: fakeDBTxn{}}
	_, err := svc.WithdrawOnChain("not-a-uuid", entities.BlockchainTRON, decimal.NewFromInt(1), "")
	if err == nil {
		t.Fatalf("expected validation error for invalid user id")
	}
}

func TestTransactionService_WithdrawOnChain_UnsupportedChain(t *testing.T) {
	svc := &TransactionService{DB: fakeDBTxn{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), "unsupported_chain", decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "unsupported blockchain") {
		t.Fatalf("expected unsupported blockchain error")
	}
}

// TronPort stubs
type tronPortNoVault struct{}

func (tronPortNoVault) SendTransaction(from, to string, amount int64, pk string) (string, error) {
	return "HASH", nil
}
func (tronPortNoVault) HasVaultConfigured() bool   { return false }
func (tronPortNoVault) GetVaultAddress() string    { return "VAULT" }
func (tronPortNoVault) GetVaultPrivateKey() string { return "PRIV" }

func TestTransactionService_WithdrawOnChain_TronServiceNotConfigured(t *testing.T) {
	svc := &TransactionService{DB: fakeDBTxn{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "TRON service not configured") {
		t.Fatalf("expected tron service not configured error")
	}
}

func TestTransactionService_WithdrawOnChain_TronVaultNotConfigured(t *testing.T) {
	svc := &TransactionService{DB: fakeDBTxn{}, Tron: tronPortNoVault{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainTRON, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "TRON vault is not configured") {
		t.Fatalf("expected tron vault not configured error")
	}
}

// Stub Ethereum gateway
type ethGateway struct{}

func (ethGateway) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: entities.BlockchainEthereum, Address: "EADDR", PublicKey: "EPUB"}, nil
}
func (ethGateway) ValidateAddress(a string) bool { return true }
func (ethGateway) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (ethGateway) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("ETHHASH"), nil
}
func (ethGateway) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (ethGateway) ChainType() entities.BlockchainType { return entities.BlockchainEthereum }

type noopOnChainRepo struct{}

func (noopOnChainRepo) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	return nil
}
func (noopOnChainRepo) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return &repositories.OnChainWallet{Address: "EADDR"}, nil
}
func (noopOnChainRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (noopOnChainRepo) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return true, nil
}

func TestTransactionService_WithdrawOnChain_EthereumMissingRegistry(t *testing.T) {
	svc := &TransactionService{DB: fakeDBTxn{}}
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "chain registry") {
		t.Fatalf("expected chain registry not configured error")
	}
}

func TestTransactionService_WithdrawOnChain_EthereumMissingRepo(t *testing.T) {
	reg := NewBlockchainRegistry(ethGateway{})
	svc := &TransactionService{DB: fakeDBTxn{}}
	svc.WithChainRegistry(reg)
	_, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(1), "")
	if err == nil || !strings.Contains(err.Error(), "on-chain wallet repository") {
		t.Fatalf("expected on-chain wallet repository not configured error")
	}
}

func TestTransactionService_WithdrawOnChain_EthereumSuccess(t *testing.T) {
	reg := NewBlockchainRegistry(ethGateway{})
	svc := &TransactionService{DB: fakeDBTxn{}}
	svc.WithChainRegistry(reg).WithOnChainWalletRepository(noopOnChainRepo{})
	// set vault env vars required
	os.Setenv("ETH_VAULT_ADDRESS", "0xVaultAddr123")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "privkey")
	resp, err := svc.WithdrawOnChain(uuid.New().String(), entities.BlockchainEthereum, decimal.NewFromInt(1), "")
	if err != nil || resp.StatusCode != 202 {
		t.Fatalf("expected ethereum withdraw success 202, got resp=%v err=%v", resp, err)
	}
}
