package gateway

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"

	"github.com/stretchr/testify/assert"
)

// NewETHGateway creates ETHGateway for testing with custom rpcURL
func NewETHGateway(rpcURL, apiKey string) *ETHGateway {
	return &ETHGateway{
		rpcURL:     rpcURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func TestETHGateway_GetTransactionHistory_InvalidAddress(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()

	txs, err := gw.GetTransactionHistory(ctx, "invalid_address", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, txs)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestETHGateway_GetTransactionHistory_ValidAddress(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()
	validAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0" // Full 40-char hex

	txs, err := gw.GetTransactionHistory(ctx, validAddr, 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, 0, len(txs)) // Returns empty slice (needs indexer)
}

func TestETHGateway_SubscribeNewBlocks_NilHandler(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()

	err := gw.SubscribeNewBlocks(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestETHGateway_SubscribeNewBlocks_OfflineMode(t *testing.T) {
	gw := NewETHGateway("", "testkey") // empty rpcURL = offline
	ctx := context.Background()

	callCount := 0
	handler := func(blockNumber int64, blockHash string, timestamp int64) error {
		callCount++
		assert.Greater(t, blockNumber, int64(0))
		assert.NotEmpty(t, blockHash)
		assert.Greater(t, timestamp, int64(0))
		return nil
	}

	err := gw.SubscribeNewBlocks(ctx, handler)

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount) // One-shot in offline mode
}

func TestETHGateway_SubscribeNewBlocks_HandlerError(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()

	expectedErr := errors.New("handler error")
	handler := func(blockNumber int64, blockHash string, timestamp int64) error {
		return expectedErr
	}

	err := gw.SubscribeNewBlocks(ctx, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestETHGateway_SubscribeNewBlocks_RPCMode_ContextCancellation(t *testing.T) {
	gw := NewETHGateway("http://localhost:8545", "testkey")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	handler := func(blockNumber int64, blockHash string, timestamp int64) error {
		callCount++
		return nil
	}

	err := gw.SubscribeNewBlocks(ctx, handler)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	// May have been called before timeout
}

func TestETHGateway_SubscribeNewTransactions_NilHandler(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()

	err := gw.SubscribeNewTransactions(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestETHGateway_SubscribeNewTransactions_InvalidAddress(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()

	handler := func(tx *entity.BlockchainTransaction) error {
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, "invalid_address", handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestETHGateway_SubscribeNewTransactions_OfflineMode(t *testing.T) {
	gw := NewETHGateway("", "testkey")
	ctx := context.Background()
	validAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	handler := func(tx *entity.BlockchainTransaction) error {
		t.Fatal("should not be called in offline mode")
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.NoError(t, err) // Returns nil immediately in offline mode
}

func TestETHGateway_SubscribeNewTransactions_RPCMode_ContextCancellation(t *testing.T) {
	gw := NewETHGateway("http://localhost:8545", "testkey")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	validAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	callCount := 0
	handler := func(tx *entity.BlockchainTransaction) error {
		callCount++
		assert.NotNil(t, tx)
		assert.NotEmpty(t, tx.TransactionHash)
		assert.Equal(t, entity.NetworkEthereum, tx.Network)
		assert.Equal(t, "confirmed", tx.Status)
		assert.Greater(t, tx.Amount.IntPart(), int64(0))
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	// May have been called before timeout
}

func TestETHGateway_SubscribeNewTransactions_HandlerError(t *testing.T) {
	gw := NewETHGateway("http://localhost:8545", "testkey")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	validAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	expectedErr := errors.New("handler processing error")
	callCount := 0
	handler := func(tx *entity.BlockchainTransaction) error {
		callCount++
		if callCount == 1 {
			// Return error on first call to test error propagation
			return expectedErr
		}
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 1, callCount)
}
