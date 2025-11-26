package gateway

import (
	"context"
	"testing"

	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"

	"github.com/stretchr/testify/assert"
)

// Test ChainType methods for all gateways
func TestETHGateway_ChainType(t *testing.T) {
	gw := NewETHGateway("", "key")
	assert.Equal(t, entity.BlockchainEthereum, gw.ChainType())
}

func TestBTCGateway_ChainType(t *testing.T) {
	gw := NewBTCGateway("")
	assert.Equal(t, entity.BlockchainBitcoin, gw.ChainType())
}

func TestSOLGateway_ChainType(t *testing.T) {
	gw := NewSOLGateway("")
	assert.Equal(t, entity.BlockchainSolana, gw.ChainType())
}

func TestTronGateway_ChainType(t *testing.T) {
	gw := &TronGateway{}
	assert.Equal(t, entity.BlockchainTron, gw.ChainType())
}

// Test GetBalance error paths
func TestETHGateway_GetBalance_InvalidAddress(t *testing.T) {
	gw := NewETHGateway("", "key")
	ctx := context.Background()

	bal, err := gw.GetBalance(ctx, "invalid_address")

	assert.Error(t, err)
	assert.Equal(t, int64(0), bal)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestBTCGateway_GetBalance_InvalidAddress(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	bal, err := gw.GetBalance(ctx, "invalid_btc_address")

	assert.Error(t, err)
	assert.Equal(t, int64(0), bal)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestSOLGateway_GetBalance_InvalidAddress(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	bal, err := gw.GetBalance(ctx, "invalid_sol_address")

	assert.Error(t, err)
	assert.Equal(t, int64(0), bal)
	assert.Contains(t, err.Error(), "invalid address")
}

// Test ValidateAddress edge cases
func TestBTCGateway_ValidateAddress_EmptyString(t *testing.T) {
	gw := NewBTCGateway("")
	assert.False(t, gw.ValidateAddress(""))
}

func TestSOLGateway_ValidateAddress_TooShort(t *testing.T) {
	gw := NewSOLGateway("")
	assert.False(t, gw.ValidateAddress("short"))
}

func TestSOLGateway_ValidateAddress_TooLong(t *testing.T) {
	gw := NewSOLGateway("")
	// Endereço com mais de 64 caracteres (65 chars para falhar)
	longAddr := "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g12345678901234567890"
	assert.False(t, gw.ValidateAddress(longAddr))
}

// Test EstimateFee error paths
func TestETHGateway_EstimateFee_InvalidFromAddress(t *testing.T) {
	gw := NewETHGateway("", "key")
	ctx := context.Background()

	fee, err := gw.EstimateFee(ctx, "invalid", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", 1000)

	assert.Error(t, err)
	assert.Nil(t, fee)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestETHGateway_EstimateFee_InvalidToAddress(t *testing.T) {
	gw := NewETHGateway("", "key")
	ctx := context.Background()

	fee, err := gw.EstimateFee(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "invalid", 1000)

	assert.Error(t, err)
	assert.Nil(t, fee)
	assert.Contains(t, err.Error(), "invalid address")
}

func TestBTCGateway_EstimateFee_InvalidAddress(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	fee, err := gw.EstimateFee(ctx, "invalid", "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 1000)

	assert.Error(t, err)
	assert.Nil(t, fee)
}

func TestSOLGateway_EstimateFee_InvalidAddress(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	fee, err := gw.EstimateFee(ctx, "invalid", "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g", 1000)

	assert.Error(t, err)
	assert.Nil(t, fee)
}

// Test Broadcast error paths
func TestETHGateway_Broadcast_EmptyPrivateKey(t *testing.T) {
	gw := NewETHGateway("", "key")
	ctx := context.Background()

	txHash, err := gw.Broadcast(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", 1000, "")

	assert.Error(t, err)
	assert.Empty(t, txHash)
}

func TestETHGateway_Broadcast_InvalidAmount(t *testing.T) {
	gw := NewETHGateway("", "key")
	ctx := context.Background()

	txHash, err := gw.Broadcast(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", 0, "abcd1234")

	assert.Error(t, err)
	assert.Empty(t, txHash)
}

func TestBTCGateway_Broadcast_InvalidAddress(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	txHash, err := gw.Broadcast(ctx, "invalid", "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 1000, "privkey")

	assert.Error(t, err)
	assert.Empty(t, txHash)
}

func TestSOLGateway_Broadcast_InvalidAddress(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	txHash, err := gw.Broadcast(ctx, "invalid", "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g", 1000, "privkey")

	assert.Error(t, err)
	assert.Empty(t, txHash)
}

// Test GenerateWallet success paths (increase coverage from 80% to 100%)
func TestETHGateway_GenerateWallet_Success(t *testing.T) {
	gw := NewETHGateway("", "key")
	ctx := context.Background()

	wallet, err := gw.GenerateWallet(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.Equal(t, entity.BlockchainEthereum, wallet.Blockchain)
	assert.Greater(t, wallet.CreatedAt, int64(0))
}

func TestBTCGateway_GenerateWallet_Success(t *testing.T) {
	gw := NewBTCGateway("")
	ctx := context.Background()

	wallet, err := gw.GenerateWallet(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.Equal(t, entity.BlockchainBitcoin, wallet.Blockchain)
	assert.Greater(t, wallet.CreatedAt, int64(0))
}

func TestSOLGateway_GenerateWallet_Success(t *testing.T) {
	gw := NewSOLGateway("")
	ctx := context.Background()

	wallet, err := gw.GenerateWallet(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.Equal(t, entity.BlockchainSolana, wallet.Blockchain)
	assert.Greater(t, wallet.CreatedAt, int64(0))
}
