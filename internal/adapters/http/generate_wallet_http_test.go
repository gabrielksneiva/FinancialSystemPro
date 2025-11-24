package http

import (
	"bytes"
	"encoding/json"
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	repositories "financial-system-pro/internal/infrastructure/database"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGenerateWallet_Ethereum(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:gen_wallet_eth?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	// minimal tables
	ddl := []string{
		"CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, password TEXT NOT NULL, created_at DATETIME)",
		"CREATE TABLE on_chain_wallets (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, blockchain TEXT NOT NULL, address TEXT NOT NULL, public_key TEXT NOT NULL, encrypted_priv_key TEXT NOT NULL, created_at DATETIME, updated_at DATETIME, UNIQUE(user_id, blockchain))",
	}
	for _, stmt := range ddl {
		require.NoError(t, db.Exec(stmt).Error)
	}
	nd := &repositories.NewDatabase{DB: db, Logger: zap.NewNop()}

	user := &repositories.User{ID: uuid.New(), Email: "wallet_gen@test.local", Password: "hashed"}
	require.NoError(t, nd.Insert(user))

	ethGw := services.NewEthereumService()
	reg := services.NewBlockchainRegistry(ethGw)
	repo := services.NewOnChainWalletRepositoryAdapter(nd)
	multiSvc := services.NewMultiChainWalletService(reg, repo)

	baseHandler := &Handler{logger: NewZapLoggerAdapter(zap.NewNop())}
	h := baseHandler.WithMultiChainWalletService(multiSvc)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", user.ID.String()); return c.Next() })
	app.Post("/api/v1/wallets/generate", h.GenerateWallet)

	body := dto.GenerateWalletRequest{Chain: "ethereum"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/wallets/generate", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, errDo := app.Test(req)
	require.NoError(t, errDo)
	require.Equal(t, 201, resp.StatusCode)
}
