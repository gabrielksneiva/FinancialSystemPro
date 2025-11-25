package http

import (
	"bytes"
	"encoding/json"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// minimal mock auth middleware injecting user_id
func injectUserBTC(id string) fiber.Handler {
	return func(c *fiber.Ctx) error { c.Locals("user_id", id); return c.Next() }
}

func TestHTTPWithdrawBitcoin(t *testing.T) {
	os.Setenv("BTC_VAULT_ADDRESS", "1VaultBitcoinAddr111111111111111111")
	os.Setenv("BTC_VAULT_PRIVATE_KEY", "TEST_PRIV_KEY_BTC")

	db, err := gorm.Open(sqlite.Open("file:btc_http_withdraw?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	ddl := []string{
		"CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, password TEXT NOT NULL, created_at DATETIME)",
		"CREATE TABLE balances (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, amount NUMERIC(15,2) NOT NULL, created_at DATETIME)",
		"CREATE TABLE transactions (id TEXT PRIMARY KEY, account_id TEXT NOT NULL, amount NUMERIC(10,2), type TEXT, category TEXT, description TEXT, tron_tx_hash TEXT, tron_tx_status TEXT, onchain_tx_hash TEXT, onchain_tx_status TEXT, onchain_chain TEXT, created_at DATETIME)",
		"CREATE TABLE on_chain_wallets (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, blockchain TEXT NOT NULL, address TEXT NOT NULL, public_key TEXT NOT NULL, encrypted_priv_key TEXT NOT NULL, created_at DATETIME, updated_at DATETIME, UNIQUE(user_id, blockchain))",
	}
	for _, stmt := range ddl {
		require.NoError(t, db.Exec(stmt).Error)
	}
	nd := &repositories.NewDatabase{DB: db, Logger: zap.NewNop()}

	user := &repositories.User{ID: uuid.New(), Email: "btc_http@test.local", Password: "hashed"}
	require.NoError(t, nd.Insert(user))
	bal := &repositories.Balance{ID: uuid.New(), UserID: user.ID, Amount: decimal.NewFromInt(50)}
	require.NoError(t, nd.Insert(bal))
	w := &repositories.OnChainWallet{ID: uuid.New(), UserID: user.ID, Blockchain: string(entities.BlockchainBitcoin), Address: "1UserBitcoinAddr1111111111111111", PublicKey: "PUB", EncryptedPrivKey: "ENC"}
	require.NoError(t, db.Create(w).Error)

	btcGw := services.NewBitcoinService()
	reg := services.NewBlockchainRegistry(btcGw)
	repo := services.NewOnChainWalletRepositoryAdapter(nd)
	txSvc := services.NewTransactionService(&services.NewDatabaseAdapter{Inner: nd}, nil, nil, nil, nil, zap.NewNop()).WithChainRegistry(reg).WithOnChainWalletRepository(repo)

	h := &Handler{transactionService: txSvc, logger: zap.NewNop()}
	app := fiber.New()
	app.Use(injectUserBTC(user.ID.String()))
	app.Post("/api/withdraw", h.Withdraw)

	reqBody := dto.WithdrawRequest{Amount: "1.5", WithdrawType: "bitcoin"}
	raw, _ := json.Marshal(reqBody)
	r := httptest.NewRequest("POST", "/api/withdraw", bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	resp, errDo := app.Test(r, -1)
	require.NoError(t, errDo)
	require.Equal(t, 202, resp.StatusCode)
}
