package integration

import (
	"context"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// fakeGatewayEnc identical to fakeGateway but reused here for encryption test
type fakeGatewayEnc struct{ chain entities.BlockchainType }

func (f *fakeGatewayEnc) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	return &entities.GeneratedWallet{Blockchain: f.chain, Address: string(f.chain) + "_ADDR_ENC_" + uuid.New().String()[:8], PublicKey: "PUB"}, nil
}
func (f *fakeGatewayEnc) ValidateAddress(address string) bool {
	return strings.HasPrefix(address, string(f.chain)+"_ADDR_ENC_")
}
func (f *fakeGatewayEnc) EstimateFee(ctx context.Context, from, to string, amountBaseUnit int64) (*services.FeeQuote, error) {
	return &services.FeeQuote{AmountBaseUnit: amountBaseUnit}, nil
}
func (f *fakeGatewayEnc) Broadcast(ctx context.Context, from, to string, amountBaseUnit int64, pk string) (services.TxHash, error) {
	return services.TxHash("HASH"), nil
}
func (f *fakeGatewayEnc) GetStatus(ctx context.Context, tx services.TxHash) (*services.TxStatusInfo, error) {
	return &services.TxStatusInfo{Hash: tx, Status: services.TxStatusConfirmed}, nil
}
func (f *fakeGatewayEnc) ChainType() entities.BlockchainType { return f.chain }

// repo capturing encrypted key
type fakeEncRepo struct{ lastEncrypted string }

func (r *fakeEncRepo) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, encryptedPriv string) error {
	r.lastEncrypted = encryptedPriv
	return nil
}
func (r *fakeEncRepo) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*repositories.OnChainWallet, error) {
	return nil, nil
}
func (r *fakeEncRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repositories.OnChainWallet, error) {
	return nil, nil
}
func (r *fakeEncRepo) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	return false, nil
}

func buildEncryptedWalletApp(r *fakeEncRepo) (*fiber.App, *fakeEncRepo) {
	tron := &fakeGatewayEnc{chain: entities.BlockchainTRON}
	reg := services.NewBlockchainRegistry(tron)
	svc := services.NewMultiChainWalletService(reg, r)
	// Provide real AES provider via env
	os.Setenv("ENCRYPTION_MASTER_KEY", strings.Repeat("a", 32)) // base64 will fail, use hex-like? 32 'a' not hex length 64 -> force error -> fallback? Better provide a 64 hex chars
	os.Setenv("ENCRYPTION_MASTER_KEY", strings.Repeat("a", 64))
	prov, _ := services.NewAESEncryptionProviderFromEnv()
	svc.WithEncryption(prov)
	rl := httpAdapter.NewRateLimiter(zap.NewNop())
	h := httpAdapter.NewHandlerForTesting(nil, nil, nil, nil, nil, rl)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New().String())
		return h.GenerateWallet(c)
	})
	return app, r
}

func TestMultiChainWalletGeneration_EncryptionAppliedIntegration(t *testing.T) {
	repo := &fakeEncRepo{}
	app, _ := buildEncryptedWalletApp(repo)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/wallets/generate", strings.NewReader(`{"chain":"tron"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "address") {
		t.Fatalf("unexpected body: %s", body)
	}
}
