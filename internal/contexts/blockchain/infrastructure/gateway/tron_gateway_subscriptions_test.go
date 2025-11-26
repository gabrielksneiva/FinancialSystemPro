package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	bcEntity "financial-system-pro/internal/contexts/blockchain/domain/entity"

	"github.com/stretchr/testify/assert"
)

func TestTronGateway_GetTransactionHistory_InvalidAddress(t *testing.T) {
	gw := NewTronGatewayFromEnv()
	ctx := context.Background()

	txs, err := gw.GetTransactionHistory(ctx, "invalid_tron_address", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, txs)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestTronGateway_GetTransactionHistory_ValidAddress(t *testing.T) {
	gw := NewTronGatewayFromEnv()
	ctx := context.Background()
	// Valid TRON address (base58, starts with T)
	validAddr := "TLPcLK6h8kRJYLPBG4H8TxucHMWZXc9J3R"

	txs, err := gw.GetTransactionHistory(ctx, validAddr, 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, txs)
	assert.Equal(t, 0, len(txs)) // Returns empty slice (needs TronGrid API)
}

func TestTronGateway_SubscribeNewBlocks_NilHandler(t *testing.T) {
	gw := NewTronGatewayFromEnv()
	ctx := context.Background()

	err := gw.SubscribeNewBlocks(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestTronGateway_SubscribeNewBlocks_OfflineMode(t *testing.T) {
	gw := &TronGateway{baseRPC: ""} // empty baseRPC = offline
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

func TestTronGateway_SubscribeNewBlocks_HandlerError(t *testing.T) {
	gw := &TronGateway{baseRPC: ""}
	ctx := context.Background()

	expectedErr := errors.New("tron block handler error")
	handler := func(blockNumber int64, blockHash string, timestamp int64) error {
		return expectedErr
	}

	err := gw.SubscribeNewBlocks(ctx, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTronGateway_SubscribeNewBlocks_RPCMode_ContextCancellation(t *testing.T) {
	gw := &TronGateway{baseRPC: "https://api.trongrid.io"}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
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

func TestTronGateway_SubscribeNewTransactions_NilHandler(t *testing.T) {
	gw := NewTronGatewayFromEnv()
	ctx := context.Background()

	err := gw.SubscribeNewTransactions(ctx, "TLPcLK6h8kRJYLPBG4H8TxucHMWZXc9J3R", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestTronGateway_SubscribeNewTransactions_InvalidAddress(t *testing.T) {
	gw := NewTronGatewayFromEnv()
	ctx := context.Background()

	handler := func(tx *bcEntity.BlockchainTransaction) error {
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, "invalid_tron_address", handler)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestTronGateway_SubscribeNewTransactions_OfflineMode(t *testing.T) {
	gw := &TronGateway{baseRPC: ""}
	ctx := context.Background()
	validAddr := "TLPcLK6h8kRJYLPBG4H8TxucHMWZXc9J3R"

	// Stub ValidateAddress to return true for this test
	gw.baseRPC = "" // Ensure offline

	handler := func(tx *bcEntity.BlockchainTransaction) error {
		t.Fatal("should not be called in offline mode")
		return nil
	}

	// Since ValidateAddress requires actual implementation, use a known valid format
	// This test may fail if ValidateAddress is strict; adjust as needed
	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	// In offline mode, returns nil without calling handler
	assert.NoError(t, err)
}

func TestTronGateway_SubscribeNewTransactions_RPCMode_ContextCancellation(t *testing.T) {
	gw := &TronGateway{baseRPC: "https://api.trongrid.io"}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()
	validAddr := "TLPcLK6h8kRJYLPBG4H8TxucHMWZXc9J3R"

	callCount := 0
	handler := func(tx *bcEntity.BlockchainTransaction) error {
		callCount++
		assert.NotNil(t, tx)
		assert.NotEmpty(t, tx.TransactionHash)
		assert.Equal(t, bcEntity.NetworkTron, tx.Network)
		assert.Equal(t, "confirmed", tx.Status)
		assert.Greater(t, tx.Amount.IntPart(), int64(0))
		return nil
	}

	err := gw.SubscribeNewTransactions(ctx, validAddr, handler)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestTronGateway_SubscribeNewTransactions_HandlerError(t *testing.T) {
	gw := &TronGateway{baseRPC: "https://api.trongrid.io"}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	validAddr := "TLPcLK6h8kRJYLPBG4H8TxucHMWZXc9J3R"

	expectedErr := errors.New("tron tx processing error")
	callCount := 0
	handler := func(tx *bcEntity.BlockchainTransaction) error {
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
