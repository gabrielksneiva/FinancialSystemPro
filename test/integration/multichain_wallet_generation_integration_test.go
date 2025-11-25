package integration

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	httpAdapter "financial-system-pro/internal/adapters/http"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	dbrepo "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// fakeGateway implements BlockchainGatewayPort for testing wallet generation
type fakeGateway struct {
	chain   entities.BlockchainType
	counter int
}

func (f *fakeGateway) GenerateWallet(ctx context.Context) (*entities.GeneratedWallet, error) {
	f.counter++
	return &entities.GeneratedWallet{Blockchain: f.chain, Address: string(f.chain) + "_ADDR_" + uuid.New().String()[:8], PublicKey: "PUB"}, nil
}
func (f *fakeGateway) ValidateAddress(address string) bool {
	return strings.HasPrefix(address, string(f.chain)+"_ADDR_")
}
func (f *fakeGateway) EstimateFee(ctx context.Context, from, to string, amountBaseUnit int64) (*services.FeeQuote, error) {
	return &services.FeeQuote{AmountBaseUnit: amountBaseUnit}, nil
}
func (f *fakeGateway) Broadcast(ctx context.Context, from, to string, amountBaseUnit int64, pk string) (services.TxHash, error) {
	return services.TxHash("HASH"), nil
}
func (f *fakeGateway) GetStatus(ctx context.Context, tx services.TxHash) (*services.TxStatusInfo, error) {
	return &services.TxStatusInfo{Hash: tx, Status: services.TxStatusConfirmed}, nil
}
func (f *fakeGateway) ChainType() entities.BlockchainType { return f.chain }

// fakeOnChainRepo stores generated wallets in-memory
type fakeOnChainRepo struct {
	store map[string]*entities.GeneratedWallet
}

func newFakeRepo() *fakeOnChainRepo {
	return &fakeOnChainRepo{store: make(map[string]*entities.GeneratedWallet)}
}
func key(uid uuid.UUID, chain entities.BlockchainType) string {
	return uid.String() + "|" + string(chain)
}
func (r *fakeOnChainRepo) Save(ctx context.Context, userID uuid.UUID, info *entities.GeneratedWallet, _ string) error {
	r.store[key(userID, info.Blockchain)] = info
	return nil
}
func (r *fakeOnChainRepo) FindByUserAndChain(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (*dbrepo.OnChainWallet, error) {
	return nil, nil
}
func (r *fakeOnChainRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*dbrepo.OnChainWallet, error) {
	return nil, nil
}
func (r *fakeOnChainRepo) Exists(ctx context.Context, userID uuid.UUID, chain entities.BlockchainType) (bool, error) {
	_, ok := r.store[key(userID, chain)]
	return ok, nil
}

// testLogger implements httpAdapter.LoggerInterface using zap noop
type testLogger struct{ *zap.Logger }

func newTestLogger() *testLogger                            { return &testLogger{zap.NewNop()} }
func (l *testLogger) Info(msg string, fields ...zap.Field)  { l.Logger.Info(msg, fields...) }
func (l *testLogger) Warn(msg string, fields ...zap.Field)  { l.Logger.Warn(msg, fields...) }
func (l *testLogger) Error(msg string, fields ...zap.Field) { l.Logger.Error(msg, fields...) }
func (l *testLogger) Debug(msg string, fields ...zap.Field) { l.Logger.Debug(msg, fields...) }

// build handler with multi-chain service
func buildHandler() (*fiber.App, *services.MultiChainWalletService) {
	tron := &fakeGateway{chain: entities.BlockchainTRON}
	eth := &fakeGateway{chain: entities.BlockchainEthereum}
	reg := services.NewBlockchainRegistry(tron, eth)
	repo := newFakeRepo()
	multi := services.NewMultiChainWalletService(reg, repo)
	h := httpAdapter.NewHandlerForTesting(nil, nil, nil, nil, nil, newTestLogger(), nil, multi)
	app := fiber.New()
	app.Post("/api/v1/wallets/generate", func(c *fiber.Ctx) error {
		// inject user_id
		if uid := c.Get("X-User-ID"); uid != "" {
			c.Locals("user_id", uid)
		}
		return h.GenerateWallet(c)
	})
	return app, multi
}

func TestMultiChainWalletGeneration_TronAndEthereum(t *testing.T) {
	app, _ := buildHandler()
	userID := uuid.New().String()

	// TRON
	req1 := httptest.NewRequest(fiber.MethodPost, "/api/v1/wallets/generate", strings.NewReader(`{"chain":"tron"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-User-ID", userID)
	resp1, _ := app.Test(req1, -1)
	if resp1.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 tron got %d", resp1.StatusCode)
	}
	body1Bytes, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()
	if !strings.Contains(string(body1Bytes), "Wallet generated") || !strings.Contains(string(body1Bytes), "tron") {
		t.Fatalf("unexpected body tron: %s", body1Bytes)
	}

	// Ethereum
	req2 := httptest.NewRequest(fiber.MethodPost, "/api/v1/wallets/generate", strings.NewReader(`{"chain":"ethereum"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-User-ID", userID)
	resp2, _ := app.Test(req2, -1)
	if resp2.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 eth got %d", resp2.StatusCode)
	}
	body2Bytes, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if !strings.Contains(string(body2Bytes), "Wallet generated") || !strings.Contains(string(body2Bytes), "ethereum") {
		t.Fatalf("unexpected body eth: %s", body2Bytes)
	}

	// Duplicate TRON should fail with 400 and have duplicate message
	req3 := httptest.NewRequest(fiber.MethodPost, "/api/v1/wallets/generate", strings.NewReader(`{"chain":"tron"}`))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-User-ID", userID)
	resp3, _ := app.Test(req3, -1)
	if resp3.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 duplicate tron got %d", resp3.StatusCode)
	}
	body3, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	if !strings.Contains(string(body3), "wallet já existe") {
		t.Fatalf("expected duplicate wallet message got %s", body3)
	}
}
