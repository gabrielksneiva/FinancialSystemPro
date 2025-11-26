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

// NewSOLGateway creates SOLGateway for testing with custom rpcURL
func NewSOLGateway(rpcURL string) *SOLGateway {
	return &SOLGateway{
		rpcURL:     rpcURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func TestSOLGateway_GetTransactionHistory_InvalidAddress(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	txs, err := gw.GetTransactionHistory(ctx, "invalid_sol_address", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, txs)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestSOLGateway_GetTransactionHistory_ValidAddress(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()
	// Valid Solana address (base58, 32-44 chars)
	validAddr := "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g"

	txs, err := gw.GetTransactionHistory(ctx, validAddr, 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, 0, len(txs)) // Returns empty slice (needs RPC calls)
}

func TestSOLGateway_SubscribeNewBlocks_NilHandler(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	err := gw.SubscribeNewBlocks(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestSOLGateway_SubscribeNewBlocks_OfflineMode(t *testing.T) {
	gw := NewSOLGateway("") // empty rpcURL = offline
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

func TestSOLGateway_SubscribeNewBlocks_HandlerError(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	expectedErr := errors.New("solana block handler error")
	handler := func(blockNumber int64, blockHash string, timestamp int64) error {
		return expectedErr
	}

	err := gw.SubscribeNewBlocks(ctx, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestSOLGateway_SubscribeNewBlocks_RPCMode_ContextCancellation(t *testing.T) {
	gw := NewSOLGateway("http://localhost:8899")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
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

func TestSOLGateway_SubscribeNewTransactions_NilHandler(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	err := gw.SubscribeNewTransactions(ctx, "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestSOLGateway_SubscribeNewTransactions_InvalidAddress(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	handler := func(tx *entity.BlockchainTransaction) error {
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, "invalid_sol_address", handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestSOLGateway_SubscribeNewTransactions_OfflineMode(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()
	validAddr := "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g"

	handler := func(tx *entity.BlockchainTransaction) error {
		t.Fatal("should not be called in offline mode")
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.NoError(t, err) // Returns nil immediately in offline mode
}

func TestSOLGateway_SubscribeNewTransactions_RPCMode_ContextCancellation(t *testing.T) {
	gw := NewSOLGateway("http://localhost:8899")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	validAddr := "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g"

	callCount := 0
	handler := func(tx *entity.BlockchainTransaction) error {
		callCount++
		assert.NotNil(t, tx)
		assert.NotEmpty(t, tx.TransactionHash)
		assert.Equal(t, entity.NetworkSolana, tx.Network)
		assert.Equal(t, "confirmed", tx.Status)
		assert.Greater(t, tx.Amount.IntPart(), int64(0))
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestSOLGateway_SubscribeNewTransactions_HandlerError(t *testing.T) {
	gw := NewSOLGateway("http://localhost:8899")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	validAddr := "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g"

	expectedErr := errors.New("sol tx processing error")
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
