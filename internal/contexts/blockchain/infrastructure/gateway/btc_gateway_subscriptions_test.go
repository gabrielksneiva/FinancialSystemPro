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

// NewBTCGateway creates BTCGateway for testing with custom rpcURL
func NewBTCGateway(rpcURL string) *BTCGateway {
	return &BTCGateway{
		rpcURL:     rpcURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func TestBTCGateway_GetTransactionHistory_InvalidAddress(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	txs, err := gw.GetTransactionHistory(ctx, "invalid_btc_address", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, txs)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestBTCGateway_GetTransactionHistory_ValidAddress(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()
	// Valid BTC mainnet address (P2PKH)
	validAddr := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	txs, err := gw.GetTransactionHistory(ctx, validAddr, 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, 0, len(txs)) // Returns empty slice (needs indexer)
}

func TestBTCGateway_SubscribeNewBlocks_NilHandler(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	err := gw.SubscribeNewBlocks(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestBTCGateway_SubscribeNewBlocks_OfflineMode(t *testing.T) {
	gw := NewBTCGateway("") // empty rpcURL = offline
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

func TestBTCGateway_SubscribeNewBlocks_HandlerError(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	expectedErr := errors.New("block handler error")
	handler := func(blockNumber int64, blockHash string, timestamp int64) error {
		return expectedErr
	}

	err := gw.SubscribeNewBlocks(ctx, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestBTCGateway_SubscribeNewBlocks_RPCMode_ContextCancellation(t *testing.T) {
	gw := NewBTCGateway("http://localhost:8332")
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
}

func TestBTCGateway_SubscribeNewTransactions_NilHandler(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	err := gw.SubscribeNewTransactions(ctx, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestBTCGateway_SubscribeNewTransactions_InvalidAddress(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	handler := func(tx *entity.BlockchainTransaction) error {
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, "invalid_btc_address", handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestBTCGateway_SubscribeNewTransactions_OfflineMode(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()
	validAddr := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	handler := func(tx *entity.BlockchainTransaction) error {
		t.Fatal("should not be called in offline mode")
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.NoError(t, err) // Returns nil immediately in offline mode
}

func TestBTCGateway_SubscribeNewTransactions_RPCMode_ContextCancellation(t *testing.T) {
	gw := NewBTCGateway("http://localhost:8332")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	validAddr := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	callCount := 0
	handler := func(tx *entity.BlockchainTransaction) error {
		callCount++
		assert.NotNil(t, tx)
		assert.NotEmpty(t, tx.TransactionHash)
		assert.Equal(t, entity.NetworkBitcoin, tx.Network)
		assert.Equal(t, "confirmed", tx.Status)
		assert.Greater(t, tx.Amount.IntPart(), int64(0))
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestBTCGateway_SubscribeNewTransactions_HandlerError(t *testing.T) {
	gw := NewBTCGateway("http://localhost:8332")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	validAddr := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	expectedErr := errors.New("tx processing error")
	callCount := 0
	handler := func(tx *entity.BlockchainTransaction) error {
		callCount++
		if callCount == 1 {
			return expectedErr
		}
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 1, callCount)
}
