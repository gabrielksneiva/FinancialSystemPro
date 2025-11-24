package services

import (
	"context"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"os"
	"testing"

	"github.com/google/uuid"
)

type fakeGateway struct{ chain entities.BlockchainType }

func (f *fakeGateway) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Address: "addr-" + string(f.chain), PublicKey: "pub-" + string(f.chain), Blockchain: f.chain}, nil
}
func (f *fakeGateway) ValidateAddress(a string) bool { return true }
func (f *fakeGateway) EstimateFee(ctx context.Context, from, to string, amt int64) (*FeeQuote, error) {
	return &FeeQuote{AmountBaseUnit: amt, EstimatedFee: 1, FeeAsset: "TEST", Source: "fake"}, nil
}
func (f *fakeGateway) Broadcast(ctx context.Context, from, to string, amt int64, pk string) (TxHash, error) {
	return TxHash("hash"), nil
}
func (f *fakeGateway) GetStatus(ctx context.Context, h TxHash) (*TxStatusInfo, error) {
	return &TxStatusInfo{Hash: h, Status: TxStatusConfirmed}, nil
}
func (f *fakeGateway) ChainType() entities.BlockchainType { return f.chain }

type captureRepo struct{ encrypted string }

func (c *captureRepo) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, encryptedPrivKey string) error {
	c.encrypted = encryptedPrivKey
	return nil
}
func (c *captureRepo) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return nil, nil
}
func (c *captureRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (c *captureRepo) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return false, nil
}

func TestMultiChainWalletService_EncryptionApplied(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef") // 64 hex chars -> 32 bytes
	prov, err := NewAESEncryptionProviderFromEnv()
	if err != nil {
		t.Fatalf("erro provider: %v", err)
	}
	repo := &captureRepo{}
	reg := NewBlockchainRegistry(&fakeGateway{chain: entities.BlockchainType("ethereum")})
	svc := NewMultiChainWalletService(reg, repo).WithEncryption(prov)
	userID := uuid.New()
	_, err = svc.GenerateAndPersist(context.Background(), userID, entities.BlockchainType("ethereum"))
	if err != nil {
		t.Fatalf("erro gerar wallet: %v", err)
	}
	if repo.encrypted == "" {
		t.Fatalf("valor criptografado vazio")
	}
	if repo.encrypted == "PRIVATE_KEY_PLACEHOLDER" {
		t.Fatalf("chave não foi criptografada")
	}
}

func TestMultiChainWalletService_NoEncryption(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "")
	repo := &captureRepo{}
	reg := NewBlockchainRegistry(&fakeGateway{chain: entities.BlockchainType("tron")})
	svc := NewMultiChainWalletService(reg, repo) // usa Noop
	userID := uuid.New()
	_, err := svc.GenerateAndPersist(context.Background(), userID, entities.BlockchainType("tron"))
	if err != nil {
		t.Fatalf("erro gerar wallet: %v", err)
	}
	if repo.encrypted != "PRIVATE_KEY_PLACEHOLDER" {
		t.Fatalf("esperava placeholder sem criptografia, got %s", repo.encrypted)
	}
}
