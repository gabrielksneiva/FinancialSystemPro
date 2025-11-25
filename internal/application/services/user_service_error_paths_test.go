package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// fake user repo triggering error
type errUserRepo struct{}

func (errUserRepo) FindByEmail(ctx context.Context, email string) (*repositories.User, error) {
	return nil, fmt.Errorf("db explode")
}
func (errUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*repositories.User, error) {
	return nil, nil
}
func (errUserRepo) Save(ctx context.Context, user *repositories.User) error { return nil }

// fake password hasher failing
type failingHasher struct{}

func (failingHasher) Compare(raw, hashed string) (bool, error) { return true, nil }
func (failingHasher) Hash(raw string) (string, error)          { return "", fmt.Errorf("hash fail") }

// minimal db implementing Insert used after verification succeeds
type minimalDB struct{}

func (minimalDB) Insert(v any) error { return nil }
func (minimalDB) FindUserByField(field string, value any) (*repositories.User, error) {
	return nil, nil
}
func (minimalDB) SaveWalletInfo(userID uuid.UUID, tronAddress, enc string) error         { return nil }
func (minimalDB) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error)       { return nil, nil }
func (minimalDB) Transaction(uid uuid.UUID, amt decimal.Decimal, tt string) error        { return nil }
func (minimalDB) Balance(uid uuid.UUID) (decimal.Decimal, error)                         { return decimal.Zero, nil }
func (minimalDB) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error { return nil }

func TestUserService_VerifyExistsError(t *testing.T) {
	svc := &UserService{UserRepo: errUserRepo{}, Logger: zap.NewNop()}
	err := svc.CreateNewUser(&dto.UserRequest{Email: "x@y.com", Password: "StrongPass1"})
	if err == nil || err.Code != errors.ErrDatabaseConnection {
		t.Fatalf("expected database error path")
	}
}

func TestUserService_HashFailPath(t *testing.T) {
	svc := NewUserService(minimalDB{}, zap.NewNop(), nil).WithPasswordHasher(failingHasher{})
	err := svc.CreateNewUser(&dto.UserRequest{Email: "z@y.com", Password: "StrongPass1"})
	if err == nil || err.Code != errors.ErrInternal {
		t.Fatalf("expected internal error from hashing")
	}
}

// encryption provider failing
type failingEncryption struct{}

func (failingEncryption) Encrypt(plain string) (string, error) { return "", fmt.Errorf("enc fail") }
func (failingEncryption) Decrypt(enc string) (string, error)   { return "", fmt.Errorf("dec fail") }

// stub gateway
type okGateway struct{ chain entities.BlockchainType }

func (g okGateway) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: g.chain, Address: "A", PublicKey: "P"}, nil
}
func (g okGateway) ValidateAddress(a string) bool { return true }
func (g okGateway) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt}, nil
}
func (g okGateway) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("HASH"), nil
}
func (g okGateway) GetStatus(ctx context.Context, tx TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: tx, Status: TxStatusConfirmed}, nil
}
func (g okGateway) ChainType() entities.BlockchainType { return g.chain }

// failing repo
type failingOnChainRepo struct{}

func (failingOnChainRepo) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, enc string) error {
	return fmt.Errorf("save fail")
}
func (failingOnChainRepo) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return nil, nil
}
func (failingOnChainRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (failingOnChainRepo) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return false, nil
}

func TestMultiChainWalletService_EncryptionFail(t *testing.T) {
	reg := NewBlockchainRegistry(okGateway{chain: entities.BlockchainTRON})
	mc := NewMultiChainWalletService(reg, failingOnChainRepo{}).WithEncryption(failingEncryption{})
	_, err := mc.GenerateAndPersist(context.Background(), uuid.New(), entities.BlockchainTRON)
	if err == nil || !strings.Contains(err.Error(), "criptografar") {
		t.Fatalf("expected encryption error path")
	}
}

func TestMultiChainWalletService_SaveFail(t *testing.T) {
	reg := NewBlockchainRegistry(okGateway{chain: entities.BlockchainTRON})
	mc := NewMultiChainWalletService(reg, failingOnChainRepo{})
	_, err := mc.GenerateAndPersist(context.Background(), uuid.New(), entities.BlockchainTRON)
	if err == nil || !strings.Contains(err.Error(), "save fail") {
		t.Fatalf("expected save fail path")
	}
}
