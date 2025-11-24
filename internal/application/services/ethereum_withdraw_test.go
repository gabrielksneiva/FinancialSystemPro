package services

import (
	"financial-system-pro/internal/domain/entities"
	repositories "financial-system-pro/internal/infrastructure/database"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestWithdrawOnChain_Ethereum basic broadcast flow
func TestWithdrawOnChain_Ethereum(t *testing.T) {
	os.Setenv("ETH_VAULT_ADDRESS", "0x0000000000000000000000000000000000000001")
	os.Setenv("ETH_VAULT_PRIVATE_KEY", "TEST_PRIV_KEY")

	db, err := gorm.Open(sqlite.Open("file:eth_withdraw_uniq?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	// criar tabelas manualmente para SQLite evitando gen_random_uuid()
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

	// create user & balance
	user := &repositories.User{ID: uuid.New(), Email: "eth_withdraw@test.local", Password: "hashed"}
	require.NoError(t, nd.Insert(user))
	bal := &repositories.Balance{ID: uuid.New(), UserID: user.ID, Amount: decimal.NewFromInt(100)}
	require.NoError(t, nd.Insert(bal))

	// create ethereum wallet record
	w := &repositories.OnChainWallet{ID: uuid.New(), UserID: user.ID, Blockchain: string(entities.BlockchainEthereum), Address: "0x00000000000000000000000000000000000000aa", PublicKey: "PUB", EncryptedPrivKey: "ENC"}
	require.NoError(t, db.Create(w).Error)

	// setup services
	ethGw := NewEthereumService()
	reg := NewBlockchainRegistry(ethGw)
	repo := NewOnChainWalletRepositoryAdapter(nd)
	svc := NewTransactionService(&NewDatabaseAdapter{Inner: nd}, nil, nil, nil, nil, zap.NewNop()).WithChainRegistry(reg).WithOnChainWalletRepository(repo)

	resp, err := svc.WithdrawOnChain(user.ID.String(), entities.BlockchainEthereum, decimal.NewFromFloat(1.25), "")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 202, resp.StatusCode)
	body := resp.Body.(map[string]interface{})
	require.Equal(t, "broadcast_success", body["status"])
	require.Contains(t, body["explorer_url"], "https://etherscan.io/tx/")
	require.Equal(t, "1.25", body["amount"])
}
