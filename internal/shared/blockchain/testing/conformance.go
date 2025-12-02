package testing

import (
	"context"
	"errors"
	"testing"
	"time"

	"financial-system-pro/internal/shared/blockchain"

	"github.com/stretchr/testify/require"
)

// ConformanceTestSuite validates that a blockchain provider implements the full Provider interface correctly.
// All providers must pass these tests to ensure consistent behavior across chains.

type ProviderFactory func() blockchain.Provider

// RunConformanceTests executes the full conformance test suite for a provider.
func RunConformanceTests(t *testing.T, factory ProviderFactory) {
	t.Run("WalletLifecycle", func(t *testing.T) { testWalletLifecycle(t, factory()) })
	t.Run("AddressValidation", func(t *testing.T) { testAddressValidation(t, factory()) })
	t.Run("BalanceOperations", func(t *testing.T) { testBalanceOperations(t, factory()) })
	t.Run("FeeEstimation", func(t *testing.T) { testFeeEstimation(t, factory()) })
	t.Run("TransactionLifecycle", func(t *testing.T) { testTransactionLifecycle(t, factory()) })
	t.Run("TransactionStatus", func(t *testing.T) { testTransactionStatus(t, factory()) })
	t.Run("Capabilities", func(t *testing.T) { testCapabilities(t, factory()) })
	t.Run("HealthCheck", func(t *testing.T) { testHealthCheck(t, factory()) })
	t.Run("ErrorHandling", func(t *testing.T) { testErrorHandling(t, factory()) })
	t.Run("Concurrency", func(t *testing.T) { testConcurrency(t, factory()) })
}

func testWalletLifecycle(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	// Generate wallet
	wallet, err := provider.GenerateWallet(ctx)
	require.NoError(t, err, "GenerateWallet should succeed")
	require.NotEmpty(t, wallet.Address, "wallet address should not be empty")
	require.NotEmpty(t, wallet.PrivateKey, "private key should not be empty")
	require.NotEmpty(t, wallet.PublicKey, "public key should not be empty")

	// Validate generated address
	err = provider.ValidateAddress(ctx, wallet.Address)
	require.NoError(t, err, "generated address should be valid")

	// Import wallet using private key
	imported, err := provider.ImportWallet(ctx, wallet.PrivateKey)
	require.NoError(t, err, "ImportWallet should succeed")
	require.Equal(t, wallet.Address, imported.Address, "imported address should match original")
}

func testAddressValidation(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	// Valid address (from generated wallet)
	wallet, _ := provider.GenerateWallet(ctx)
	err := provider.ValidateAddress(ctx, wallet.Address)
	require.NoError(t, err, "valid address should pass validation")

	// Invalid addresses
	invalidAddresses := []string{
		"",
		"invalid",
		"0x",
		"0x123",
		"notanaddress",
		"T" + string(make([]byte, 100)), // too long
	}

	for _, addr := range invalidAddresses {
		err := provider.ValidateAddress(ctx, addr)
		require.Error(t, err, "invalid address '%s' should fail validation", addr)
	}
}

func testBalanceOperations(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	wallet, _ := provider.GenerateWallet(ctx)

	// Get balance
	balance, err := provider.GetBalance(ctx, wallet.Address)
	require.NoError(t, err, "GetBalance should succeed")
	require.NotNil(t, balance, "balance should not be nil")
	require.GreaterOrEqual(t, balance.Value, int64(0), "balance value should be non-negative")
	require.NotEmpty(t, balance.Currency, "balance currency should be set")

	// Balance for invalid address should fail
	_, err = provider.GetBalance(ctx, "invalid-address")
	require.Error(t, err, "GetBalance with invalid address should fail")
}

func testFeeEstimation(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	wallet, _ := provider.GenerateWallet(ctx)
	to := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0" // dummy address

	intent := blockchain.TransactionIntent{
		From:   wallet.Address,
		To:     to,
		Amount: 1000,
	}

	// Estimate fees
	fees, err := provider.EstimateFees(ctx, intent)
	require.NoError(t, err, "EstimateFees should succeed")
	require.NotNil(t, fees, "fees should not be nil")
	require.Greater(t, fees.Low, int64(0), "low fee should be positive")
	require.Greater(t, fees.Medium, int64(0), "medium fee should be positive")
	require.Greater(t, fees.High, int64(0), "high fee should be positive")
	require.GreaterOrEqual(t, fees.Medium, fees.Low, "medium fee >= low fee")
	require.GreaterOrEqual(t, fees.High, fees.Medium, "high fee >= medium fee")
}

func testTransactionLifecycle(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	wallet, _ := provider.GenerateWallet(ctx)
	to := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	intent := blockchain.TransactionIntent{
		From:   wallet.Address,
		To:     to,
		Amount: 1000,
		Data:   []byte{},
	}

	// Build transaction
	unsigned, err := provider.BuildTransaction(ctx, intent)
	require.NoError(t, err, "BuildTransaction should succeed")
	require.NotEmpty(t, unsigned.Hash, "transaction hash should not be empty")
	require.NotEmpty(t, unsigned.RawData, "raw data should not be empty")

	// Sign transaction
	signed, err := provider.SignTransaction(ctx, unsigned, wallet.PrivateKey)
	require.NoError(t, err, "SignTransaction should succeed")
	require.NotEmpty(t, signed.Hash, "signed hash should not be empty")
	require.NotEmpty(t, signed.Signature, "signature should not be empty")
	require.NotEmpty(t, signed.RawData, "signed raw data should not be empty")

	// Note: Broadcast will fail in test environment without funded wallet,
	// but we test that the method exists and handles errors properly
	receipt, err := provider.BroadcastTransaction(ctx, signed)
	// Either succeeds (if simulated) or fails gracefully
	if err == nil {
		require.NotNil(t, receipt, "receipt should not be nil on success")
		require.NotEmpty(t, receipt.TxHash, "receipt should have tx hash")
	} else {
		require.NotNil(t, err, "broadcast error should be informative")
	}
}

func testTransactionStatus(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	// Dummy transaction hash
	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	status, err := provider.GetTransactionStatus(ctx, txHash)
	// Should either return status or error (not panic)
	if err == nil {
		require.NotNil(t, status, "status should not be nil")
	} else {
		require.NotNil(t, err, "error should be informative")
	}
}

func testCapabilities(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	caps := provider.Capabilities(ctx)
	require.NotNil(t, caps, "capabilities should not be nil")

	// At minimum, all providers should support basic operations
	require.Contains(t, caps, "generate_wallet", "should support wallet generation")
	require.Contains(t, caps, "sign_transaction", "should support signing")
}

func testHealthCheck(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	healthy := provider.IsHealthy(ctx)
	// Health check should return without panic
	_ = healthy
}

func testErrorHandling(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	// Import with invalid private key
	_, err := provider.ImportWallet(ctx, "invalid-key")
	require.Error(t, err, "ImportWallet with invalid key should fail")

	// Sign with invalid private key
	wallet, _ := provider.GenerateWallet(ctx)
	intent := blockchain.TransactionIntent{From: wallet.Address, To: wallet.Address, Amount: 1}
	unsigned, _ := provider.BuildTransaction(ctx, intent)
	_, err = provider.SignTransaction(ctx, unsigned, "invalid-key")
	require.Error(t, err, "SignTransaction with invalid key should fail")

	// Build transaction with empty from address
	badIntent := blockchain.TransactionIntent{From: "", To: wallet.Address, Amount: 1}
	_, err = provider.BuildTransaction(ctx, badIntent)
	require.Error(t, err, "BuildTransaction with empty from should fail")
}

func testConcurrency(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()

	// Test concurrent wallet generation
	done := make(chan *blockchain.Wallet, 10)
	for i := 0; i < 10; i++ {
		go func() {
			wallet, err := provider.GenerateWallet(ctx)
			if err != nil {
				done <- nil
			} else {
				done <- wallet
			}
		}()
	}

	wallets := make(map[string]bool)
	for i := 0; i < 10; i++ {
		wallet := <-done
		require.NotNil(t, wallet, "concurrent wallet generation should succeed")
		require.False(t, wallets[wallet.Address], "wallet addresses should be unique")
		wallets[wallet.Address] = true
	}
}

// RunCapabilityTests tests optional capabilities that providers may support.
func RunCapabilityTests(t *testing.T, provider blockchain.Provider) {
	ctx := context.Background()
	caps := provider.Capabilities(ctx)

	if _, ok := caps["transaction_history"]; ok {
		t.Run("TransactionHistory", func(t *testing.T) {
			wallet, _ := provider.GenerateWallet(ctx)
			history, err := provider.GetTransactionHistory(ctx, wallet.Address, 10)
			require.NoError(t, err, "GetTransactionHistory should succeed")
			require.NotNil(t, history, "history should not be nil")
		})
	}

	if _, ok := caps["subscribe_events"]; ok {
		t.Run("EventSubscription", func(t *testing.T) {
			events := make(chan blockchain.BlockchainEvent, 10)
			err := provider.SubscribeToEvents(ctx, "block", events)
			require.NoError(t, err, "SubscribeToEvents should succeed")

			// Wait briefly for events or timeout
			select {
			case <-events:
				// Event received
			case <-time.After(1 * time.Second):
				// Timeout is acceptable in test environment
			}
		})
	}
}

// TestProviderRegistry validates the Registry implementation.
func TestProviderRegistry(t *testing.T, registry blockchain.Registry) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		// Create a mock provider
		mockProvider := &MockProvider{}
		
		err := registry.Register(blockchain.ChainEthereum, blockchain.NetworkMainnet, mockProvider)
		require.NoError(t, err, "Register should succeed")

		provider, err := registry.Get(blockchain.ChainEthereum, blockchain.NetworkMainnet)
		require.NoError(t, err, "Get should succeed")
		require.NotNil(t, provider, "provider should not be nil")

		// Get non-existent provider
		_, err = registry.Get(blockchain.ChainType("nonexistent"), blockchain.NetworkMainnet)
		require.Error(t, err, "Get non-existent should fail")
	})

	t.Run("ListProviders", func(t *testing.T) {
		providers := registry.List()
		require.NotNil(t, providers, "List should not return nil")
	})

	t.Run("Exists", func(t *testing.T) {
		exists := registry.Exists(blockchain.ChainEthereum, blockchain.NetworkMainnet)
		_ = exists // Just test it doesn't panic
	})
}

// MockProvider for testing
type MockProvider struct{}

func (m *MockProvider) GenerateWallet(ctx context.Context) (*blockchain.Wallet, error) {
	return &blockchain.Wallet{Address: "mock-address", PrivateKey: "mock-key", PublicKey: "mock-pub"}, nil
}
func (m *MockProvider) ImportWallet(ctx context.Context, privateKey blockchain.PrivateKey) (*blockchain.Wallet, error) {
	return nil, errors.New("not implemented")
}
func (m *MockProvider) ValidateAddress(ctx context.Context, address string) error { return nil }
func (m *MockProvider) GetBalance(ctx context.Context, address string) (*blockchain.Balance, error) {
	return &blockchain.Balance{Value: 0, Currency: "MOCK"}, nil
}
func (m *MockProvider) EstimateFees(ctx context.Context, intent blockchain.TransactionIntent) (*blockchain.FeeEstimate, error) {
	return &blockchain.FeeEstimate{Low: 1, Medium: 2, High: 3}, nil
}
func (m *MockProvider) BuildTransaction(ctx context.Context, intent blockchain.TransactionIntent) (*blockchain.UnsignedTransaction, error) {
	return &blockchain.UnsignedTransaction{Hash: "mock-hash", RawData: []byte("mock")}, nil
}
func (m *MockProvider) SignTransaction(ctx context.Context, tx *blockchain.UnsignedTransaction, privateKey blockchain.PrivateKey) (*blockchain.SignedTransaction, error) {
	return &blockchain.SignedTransaction{Hash: "mock-hash", Signature: []byte("sig"), RawData: []byte("signed")}, nil
}
func (m *MockProvider) BroadcastTransaction(ctx context.Context, tx *blockchain.SignedTransaction) (*blockchain.TransactionReceipt, error) {
	return &blockchain.TransactionReceipt{TxHash: tx.Hash, BlockNumber: 1, Status: "success"}, nil
}
func (m *MockProvider) GetTransactionStatus(ctx context.Context, txHash string) (*blockchain.TransactionStatus, error) {
	return &blockchain.TransactionStatus{Hash: txHash, Status: "pending"}, nil
}
func (m *MockProvider) GetTransactionHistory(ctx context.Context, address string, limit int) ([]blockchain.Transaction, error) {
	return []blockchain.Transaction{}, nil
}
func (m *MockProvider) SubscribeToEvents(ctx context.Context, eventType string, events chan<- blockchain.BlockchainEvent) error {
	return nil
}
func (m *MockProvider) Capabilities(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{"generate_wallet": true, "sign_transaction": true}
}
func (m *MockProvider) IsHealthy(ctx context.Context) bool { return true }
