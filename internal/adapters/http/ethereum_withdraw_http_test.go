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
func injectUser(id string) fiber.Handler {
	return func(c *fiber.Ctx) error { c.Locals("user_id", id); return c.Next() }
}

func TestHTTPWithdrawEthereum(t *testing.T) {
	os.Setenv("ETH_VAULT_ADDRESS", "0x0000000000000000000000000000000000000001")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "TEST_PRIV_KEY")

	db, err := gorm.Open(sqlite.Open("file:eth_http_withdraw?mode=memory&cache=shared"), &gorm.Config{})
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

	user := &repositories.User{ID: uuid.New(), Email: "eth_http@test.local", Password: "hashed"}
	require.NoError(t, nd.Insert(user))
	bal := &repositories.Balance{ID: uuid.New(), UserID: user.ID, Amount: decimal.NewFromInt(50)}
	require.NoError(t, nd.Insert(bal))
	w := &repositories.OnChainWallet{ID: uuid.New(), UserID: user.ID, Blockchain: string(entities.BlockchainEthereum), Address: "0x00000000000000000000000000000000000000aa", PublicKey: "PUB", EncryptedPrivKey: "ENC"}
	require.NoError(t, db.Create(w).Error)

	ethGw := services.NewEthereumService()
	reg := services.NewBlockchainRegistry(ethGw)
	repo := services.NewOnChainWalletRepositoryAdapter(nd)
	txSvc := services.NewTransactionService(&services.NewDatabaseAdapter{Inner: nd}, nil, nil, nil, nil, zap.NewNop()).WithChainRegistry(reg).WithOnChainWalletRepository(repo)

	h := &Handler{transactionService: txSvc, logger: zap.NewNop()}
	app := fiber.New()
	app.Use(injectUser(user.ID.String()))
	app.Post("/api/withdraw", h.Withdraw)

	reqBody := dto.WithdrawRequest{Amount: "1.5", WithdrawType: "ethereum"}
	raw, _ := json.Marshal(reqBody)
	r := httptest.NewRequest("POST", "/api/withdraw", bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	resp, errDo := app.Test(r, -1)
	require.NoError(t, errDo)
	require.Equal(t, 202, resp.StatusCode)
}
