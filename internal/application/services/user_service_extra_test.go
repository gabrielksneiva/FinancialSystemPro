package services

import (
	"context"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// fake DB implementing minimal Insert + FindUserByField for user creation path
type fakeDB struct{ insertedEmail string }

func (f *fakeDB) FindUserByField(field string, value any) (*repositories.User, error) {
	return nil, nil
}
func (f *fakeDB) Insert(value any) error {
	if u, ok := value.(*repositories.User); ok {
		f.insertedEmail = u.Email
	}
	return nil
}
func (f *fakeDB) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error   { return nil }
func (f *fakeDB) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) { return nil, nil }
func (f *fakeDB) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return nil
}
func (f *fakeDB) Balance(userID uuid.UUID) (decimal.Decimal, error)                      { return decimal.Zero, nil }
func (f *fakeDB) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error { return nil }

// minimal wallet manager
type fakeWalletManager struct{}

func (f fakeWalletManager) GenerateWallet() (*entities.WalletInfo, error) {
	return &entities.WalletInfo{Address: "ADDR", EncryptedPrivKey: "ENC", Blockchain: entities.BlockchainTRON}, nil
}
func (f fakeWalletManager) ValidateAddress(address string) bool { return true }
func (f fakeWalletManager) GetBlockchainType() entities.BlockchainType {
	return entities.BlockchainTRON
}

func TestUserService_WithMultiChainWalletService(t *testing.T) {
	db := &fakeDB{}
	svc := NewUserService(db, zap.NewNop(), fakeWalletManager{})
	// inject multi-chain service with registry having TRON only (should attempt generation)
	tronGateway := &stubGateway{chain: entities.BlockchainTRON}
	reg := NewBlockchainRegistry(tronGateway)
	repo := &stubOnChainRepo{}
	mc := NewMultiChainWalletService(reg, repo)
	svc.WithMultiChainWalletService(mc)
	req := &dto.UserRequest{Email: "x@y.com", Password: "StrongPass1"}
	err := svc.CreateNewUser(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.insertedEmail != "x@y.com" {
		t.Fatalf("user not inserted")
	}
	if repo.saveCalls == 0 {
		t.Fatalf("expected multi-chain wallet save to be invoked")
	}
}

// stub gateway & repo for multi-chain
type stubGateway struct{ chain entities.BlockchainType }

func (s *stubGateway) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: s.chain, Address: "CADDR", PublicKey: "PUB"}, nil
}
func (s *stubGateway) ValidateAddress(a string) bool { return true }
func (s *stubGateway) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (s *stubGateway) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("HASH"), nil
}
func (s *stubGateway) GetStatus(ctx context.Context, tx TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: tx, Status: TxStatusConfirmed}, nil
}
func (s *stubGateway) ChainType() entities.BlockchainType { return s.chain }

type stubOnChainRepo struct{ saveCalls int }

func (s *stubOnChainRepo) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	s.saveCalls++
	return nil
}
func (s *stubOnChainRepo) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return nil, nil
}
func (s *stubOnChainRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (s *stubOnChainRepo) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return false, nil
}
