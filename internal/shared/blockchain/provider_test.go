package blockchain

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider é um provider mock para testes
type MockProvider struct {
	chainType    ChainType
	networkType  NetworkType
	isHealthy    bool
	capabilities *ProviderCapabilities
}

func NewMockProvider(chainType ChainType, networkType NetworkType) *MockProvider {
	return &MockProvider{
		chainType:   chainType,
		networkType: networkType,
		isHealthy:   true,
		capabilities: &ProviderCapabilities{
			SupportsSmartContracts:   true,
			SupportsTokens:           true,
			SupportsSubscriptions:    true,
			NativeTokenDecimals:      18,
			MinConfirmationsRequired: 12,
			AverageBlockTime:         12 * time.Second,
		},
	}
}

func (m *MockProvider) ChainType() ChainType {
	return m.chainType
}

func (m *MockProvider) NetworkType() NetworkType {
	return m.networkType
}

func (m *MockProvider) IsHealthy(ctx context.Context) bool {
	return m.isHealthy
}

func (m *MockProvider) GenerateWallet(ctx context.Context) (*Wallet, error) {
	return &Wallet{
		Address:    "0xmock" + string(m.chainType),
		PublicKey:  "mock_public_key",
		PrivateKey: "mock_private_key",
		ChainType:  m.chainType,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockProvider) ValidateAddress(address string) error {
	if address == "" {
		return ErrInvalidAddress
	}
	return nil
}

func (m *MockProvider) ImportWallet(ctx context.Context, privateKey string) (*Wallet, error) {
	return &Wallet{
		Address:    "0ximported",
		PublicKey:  "imported_public_key",
		PrivateKey: privateKey,
		ChainType:  m.chainType,
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockProvider) GetBalance(ctx context.Context, address string) (*Balance, error) {
	return &Balance{
		Address:       address,
		Amount:        decimal.NewFromInt(100),
		AmountRaw:     "100000000000000000000",
		Currency:      "MOCK",
		Decimals:      18,
		BlockNumber:   1000,
		LastUpdatedAt: time.Now(),
	}, nil
}

func (m *MockProvider) EstimateFee(ctx context.Context, intent *TransactionIntent) (*FeeEstimate, error) {
	return &FeeEstimate{
		ChainType:   m.chainType,
		Low:         decimal.NewFromFloat(0.001),
		Medium:      decimal.NewFromFloat(0.002),
		High:        decimal.NewFromFloat(0.003),
		Currency:    "MOCK",
		EstimatedAt: time.Now(),
	}, nil
}

func (m *MockProvider) BuildTransaction(ctx context.Context, intent *TransactionIntent) (*UnsignedTransaction, error) {
	return &UnsignedTransaction{
		ChainType: m.chainType,
		From:      intent.From,
		To:        intent.To,
		Amount:    intent.Amount,
		Fee:       decimal.NewFromFloat(0.002),
		Nonce:     1,
		Data:      intent.Data,
		RawData:   make(map[string]interface{}),
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockProvider) SignTransaction(ctx context.Context, tx *UnsignedTransaction, privateKey *PrivateKey) (*SignedTransaction, error) {
	return &SignedTransaction{
		ChainType: m.chainType,
		RawTx:     "0xsigned_tx_data",
		TxHash:    "0xmock_tx_hash",
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Signature: "mock_signature",
		RawData:   make(map[string]interface{}),
		SignedAt:  time.Now(),
	}, nil
}

func (m *MockProvider) BroadcastTransaction(ctx context.Context, tx *SignedTransaction) (*TransactionReceipt, error) {
	return &TransactionReceipt{
		TxHash:        tx.TxHash,
		ChainType:     m.chainType,
		Status:        TxStatusPending,
		From:          tx.From,
		To:            tx.To,
		Amount:        tx.Amount,
		Fee:           tx.Fee,
		Confirmations: 0,
		BroadcastAt:   time.Now(),
	}, nil
}

func (m *MockProvider) GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	return &TransactionStatus{
		TxHash:        txHash,
		Status:        TxStatusConfirmed,
		Confirmations: 12,
		BlockNumber:   func() *int64 { n := int64(1000); return &n }(),
	}, nil
}

func (m *MockProvider) GetTransactionHistory(ctx context.Context, address string, opts *PaginationOptions) (*TransactionHistory, error) {
	return &TransactionHistory{
		Address: address,
		Transactions: []HistoricalTransaction{
			{
				TxHash:        "0xhist1",
				From:          address,
				To:            "0xto",
				Amount:        decimal.NewFromInt(10),
				Fee:           decimal.NewFromFloat(0.002),
				Status:        TxStatusConfirmed,
				BlockNumber:   999,
				Confirmations: 100,
				Timestamp:     time.Now().Add(-24 * time.Hour),
				Type:          "sent",
			},
		},
		Total:   1,
		HasMore: false,
	}, nil
}

func (m *MockProvider) SubscribeNewBlocks(ctx context.Context, handler BlockHandler) error {
	return nil
}

func (m *MockProvider) SubscribeNewTransactions(ctx context.Context, filter *TxFilter, handler TxHandler) error {
	return nil
}

func (m *MockProvider) UnsubscribeAll(ctx context.Context) error {
	return nil
}

func (m *MockProvider) GetCapabilities() *ProviderCapabilities {
	return m.capabilities
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	t.Run("register provider successfully", func(t *testing.T) {
		provider := NewMockProvider(ChainEthereum, NetworkMainnet)
		err := registry.Register(provider)
		require.NoError(t, err)

		assert.True(t, registry.Exists(ChainEthereum))
		assert.Equal(t, 1, registry.Count())
	})

	t.Run("reject duplicate registration", func(t *testing.T) {
		provider1 := NewMockProvider(ChainBitcoin, NetworkMainnet)
		provider2 := NewMockProvider(ChainBitcoin, NetworkTestnet)

		err1 := registry.Register(provider1)
		require.NoError(t, err1)

		err2 := registry.Register(provider2)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "already registered")
	})

	t.Run("reject nil provider", func(t *testing.T) {
		err := registry.Register(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	t.Run("get existing provider", func(t *testing.T) {
		provider := NewMockProvider(ChainSolana, NetworkMainnet)
		err := registry.Register(provider)
		require.NoError(t, err)

		retrieved, err := registry.Get(ChainSolana)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, ChainSolana, retrieved.ChainType())
	})

	t.Run("fail to get non-existent provider", func(t *testing.T) {
		_, err := registry.Get(ChainTON)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no provider registered")
	})
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	providers := []Provider{
		NewMockProvider(ChainEthereum, NetworkMainnet),
		NewMockProvider(ChainBitcoin, NetworkMainnet),
		NewMockProvider(ChainSolana, NetworkMainnet),
	}

	for _, p := range providers {
		err := registry.Register(p)
		require.NoError(t, err)
	}

	list := registry.List()
	assert.Len(t, list, 3)
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	provider := NewMockProvider(ChainTron, NetworkMainnet)

	err := registry.Register(provider)
	require.NoError(t, err)
	assert.True(t, registry.Exists(ChainTron))

	err = registry.Unregister(ChainTron)
	require.NoError(t, err)
	assert.False(t, registry.Exists(ChainTron))
}

func TestProviderContract_WalletOperations(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider(ChainEthereum, NetworkMainnet)

	t.Run("generate wallet", func(t *testing.T) {
		wallet, err := provider.GenerateWallet(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, wallet.Address)
		assert.NotEmpty(t, wallet.PublicKey)
		assert.Equal(t, ChainEthereum, wallet.ChainType)
	})

	t.Run("validate address", func(t *testing.T) {
		err := provider.ValidateAddress("0x123")
		assert.NoError(t, err)

		err = provider.ValidateAddress("")
		assert.Error(t, err)
	})

	t.Run("import wallet", func(t *testing.T) {
		wallet, err := provider.ImportWallet(ctx, "private_key_123")
		require.NoError(t, err)
		assert.NotEmpty(t, wallet.Address)
	})
}

func TestProviderContract_TransactionFlow(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider(ChainEthereum, NetworkMainnet)

	t.Run("full transaction flow", func(t *testing.T) {
		// 1. Estimate fee
		intent := &TransactionIntent{
			From:   "0xfrom",
			To:     "0xto",
			Amount: decimal.NewFromInt(10),
		}

		feeEstimate, err := provider.EstimateFee(ctx, intent)
		require.NoError(t, err)
		assert.True(t, feeEstimate.Medium.GreaterThan(decimal.Zero))

		// 2. Build transaction
		unsignedTx, err := provider.BuildTransaction(ctx, intent)
		require.NoError(t, err)
		assert.Equal(t, intent.From, unsignedTx.From)
		assert.Equal(t, intent.To, unsignedTx.To)

		// 3. Sign transaction
		privateKey := &PrivateKey{Raw: []byte("test_key")}
		signedTx, err := provider.SignTransaction(ctx, unsignedTx, privateKey)
		require.NoError(t, err)
		assert.NotEmpty(t, signedTx.TxHash)
		assert.NotEmpty(t, signedTx.Signature)

		// 4. Broadcast transaction
		receipt, err := provider.BroadcastTransaction(ctx, signedTx)
		require.NoError(t, err)
		assert.Equal(t, signedTx.TxHash, receipt.TxHash)

		// 5. Check status
		status, err := provider.GetTransactionStatus(ctx, receipt.TxHash)
		require.NoError(t, err)
		assert.Equal(t, receipt.TxHash, status.TxHash)
	})
}

func TestProviderContract_BalanceAndHistory(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider(ChainEthereum, NetworkMainnet)
	address := "0xtest"

	t.Run("get balance", func(t *testing.T) {
		balance, err := provider.GetBalance(ctx, address)
		require.NoError(t, err)
		assert.Equal(t, address, balance.Address)
		assert.True(t, balance.Amount.GreaterThan(decimal.Zero))
	})

	t.Run("get transaction history", func(t *testing.T) {
		opts := &PaginationOptions{Limit: 10, Offset: 0}
		history, err := provider.GetTransactionHistory(ctx, address, opts)
		require.NoError(t, err)
		assert.Equal(t, address, history.Address)
		assert.NotEmpty(t, history.Transactions)
	})
}

func TestPrivateKey_Security(t *testing.T) {
	t.Run("private key is redacted in JSON", func(t *testing.T) {
		pk := &PrivateKey{Raw: []byte("super_secret_key")}
		data, err := pk.MarshalJSON()
		require.NoError(t, err)
		assert.NotContains(t, string(data), "super_secret_key")
		assert.Contains(t, string(data), "REDACTED")
	})
}

func TestChainError(t *testing.T) {
	t.Run("chain error formatting", func(t *testing.T) {
		err := NewChainError(ChainEthereum, "broadcast", ErrBroadcastFailed)
		assert.Contains(t, err.Error(), "ethereum")
		assert.Contains(t, err.Error(), "broadcast")
		assert.Contains(t, err.Error(), "broadcast failed")
	})

	t.Run("chain error unwrap", func(t *testing.T) {
		originalErr := ErrInvalidAddress
		chainErr := NewChainError(ChainBitcoin, "validate", originalErr)
		assert.Equal(t, originalErr, chainErr.Unwrap())
	})
}
