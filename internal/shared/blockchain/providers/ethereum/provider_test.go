package ethereum_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"financial-system-pro/internal/shared/blockchain"
	"financial-system-pro/internal/shared/blockchain/providers/ethereum"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRPCServer simula um endpoint JSON-RPC Ethereum
type MockRPCServer struct {
	server   *httptest.Server
	requests []map[string]interface{}
	handler  func(method string, params []interface{}) (interface{}, error)
}

func NewMockRPCServer() *MockRPCServer {
	mock := &MockRPCServer{
		requests: make([]map[string]interface{}, 0),
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)
		mock.requests = append(mock.requests, req)

		method := req["method"].(string)
		params := []interface{}{}
		if req["params"] != nil {
			params = req["params"].([]interface{})
		}

		var result interface{}
		var err error
		if mock.handler != nil {
			result, err = mock.handler(method, params)
		} else {
			result, err = mock.defaultHandler(method, params)
		}

		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req["id"],
		}

		if err != nil {
			response["error"] = map[string]interface{}{
				"code":    -32000,
				"message": err.Error(),
			}
		} else {
			response["result"] = result
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	return mock
}

func (m *MockRPCServer) defaultHandler(method string, params []interface{}) (interface{}, error) {
	switch method {
	case "eth_getBalance":
		return "0xde0b6b3a7640000", nil // 1 ETH in wei
	case "eth_gasPrice":
		return "0x3b9aca00", nil // 1 gwei
	case "eth_estimateGas":
		return "0x5208", nil // 21000 gas
	case "eth_getTransactionCount":
		return "0x1", nil
	case "eth_chainId":
		return "0x1", nil
	case "eth_sendRawTransaction":
		return "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", nil
	case "eth_getTransactionByHash":
		return map[string]interface{}{
			"hash":             params[0],
			"blockHash":        "0xabcd1234",
			"blockNumber":      "0x12345",
			"from":             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
			"to":               "0x5AEDA56215b167893e80B4fE645BA6d5Bab767DE",
			"value":            "0xde0b6b3a7640000",
			"gas":              "0x5208",
			"gasPrice":         "0x3b9aca00",
			"nonce":            "0x1",
			"transactionIndex": "0x0",
		}, nil
	case "eth_getTransactionReceipt":
		return map[string]interface{}{
			"transactionHash":   params[0],
			"blockHash":         "0xabcd1234",
			"blockNumber":       "0x12345",
			"status":            "0x1",
			"gasUsed":           "0x5208",
			"cumulativeGasUsed": "0x5208",
		}, nil
	case "eth_blockNumber":
		return "0x12345", nil
	case "eth_getBlockByNumber":
		return map[string]interface{}{
			"number":       params[0],
			"hash":         "0xabcd1234",
			"timestamp":    "0x60000000",
			"transactions": []interface{}{},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func (m *MockRPCServer) Close() {
	m.server.Close()
}

func (m *MockRPCServer) URL() string {
	return m.server.URL
}

func TestEthereumProvider_GenerateWallet(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	wallet, err := provider.GenerateWallet(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.True(t, strings.HasPrefix(wallet.Address, "0x"))
	assert.Equal(t, 42, len(wallet.Address)) // 0x + 40 hex chars
}

func TestEthereumProvider_ValidateAddress(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	tests := []struct {
		name    string
		address string
		valid   bool
	}{
		{"Valid address", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", true},
		{"Valid lowercase", "0x742d35cc6634c0532925a3b844bc9e7595f0beb1", true},
		{"Missing 0x (accepted)", "742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", true},
		{"Too short", "0x742d35Cc", false},
		{"Invalid chars", "0xZZZZ35Cc6634C0532925a3b844Bc9e7595f0bEb1", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateAddress(tt.address)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEthereumProvider_GetBalance(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	balance, err := provider.GetBalance(context.Background(), "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1")
	require.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, "ETH", balance.Currency)
	assert.True(t, balance.Amount.GreaterThan(decimal.Zero))
}

func TestEthereumProvider_EstimateFee(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	intent := blockchain.TransactionIntent{
		From:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		To:     "0x5AEDA56215b167893e80B4fE645BA6d5Bab767DE",
		Amount: decimal.NewFromFloat(1.0),
	}

	estimate, err := provider.EstimateFee(context.Background(), &intent)
	require.NoError(t, err)
	assert.NotNil(t, estimate)
	assert.True(t, estimate.Low.GreaterThan(decimal.Zero))
	assert.NotNil(t, estimate.GasLimit)
	assert.True(t, *estimate.GasLimit > 0)
}

func TestEthereumProvider_BuildTransaction(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	intent := blockchain.TransactionIntent{
		From:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		To:     "0x5AEDA56215b167893e80B4fE645BA6d5Bab767DE",
		Amount: decimal.NewFromFloat(1.0),
	}

	unsigned, err := provider.BuildTransaction(context.Background(), &intent)
	require.NoError(t, err)
	assert.NotNil(t, unsigned)
	assert.Equal(t, intent.From, unsigned.From)
	assert.Equal(t, intent.To, unsigned.To)
	assert.True(t, unsigned.Amount.Equal(intent.Amount))
	assert.NotEmpty(t, unsigned.RawData)
}

func TestEthereumProvider_SignTransaction(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	// Generate wallet for signing
	wallet, err := provider.GenerateWallet(context.Background())
	require.NoError(t, err)

	// Decode private key hex -> raw bytes (remove 0x)
	pkBytes, err := hex.DecodeString(strings.TrimPrefix(wallet.PrivateKey, "0x"))
	require.NoError(t, err)
	privKey := blockchain.PrivateKey{Raw: pkBytes}

	intent := blockchain.TransactionIntent{
		From:   wallet.Address,
		To:     "0x5AEDA56215b167893e80B4fE645BA6d5Bab767DE",
		Amount: decimal.NewFromFloat(1.0),
	}

	unsigned, err := provider.BuildTransaction(context.Background(), &intent)
	require.NoError(t, err)

	signed, err := provider.SignTransaction(context.Background(), unsigned, &privKey)
	require.NoError(t, err)
	assert.NotNil(t, signed)
	assert.NotEmpty(t, signed.RawTx)
	assert.NotEmpty(t, signed.TxHash)
}

func TestEthereumProvider_BroadcastTransaction(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	wallet, err := provider.GenerateWallet(context.Background())
	require.NoError(t, err)
	pkBytes, err := hex.DecodeString(strings.TrimPrefix(wallet.PrivateKey, "0x"))
	require.NoError(t, err)
	privKey := blockchain.PrivateKey{Raw: pkBytes}

	intent := blockchain.TransactionIntent{
		From:   wallet.Address,
		To:     "0x5AEDA56215b167893e80B4fE645BA6d5Bab767DE",
		Amount: decimal.NewFromFloat(1.0),
	}

	unsigned, err := provider.BuildTransaction(context.Background(), &intent)
	require.NoError(t, err)

	signed, err := provider.SignTransaction(context.Background(), unsigned, &privKey)
	require.NoError(t, err)

	receipt, err := provider.BroadcastTransaction(context.Background(), signed)
	require.NoError(t, err)
	assert.NotEmpty(t, receipt.TxHash)
	assert.True(t, strings.HasPrefix(receipt.TxHash, "0x"))
}

func TestEthereumProvider_GetTransactionStatus(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	status, err := provider.GetTransactionStatus(context.Background(), txHash)
	require.NoError(t, err)
	assert.NotNil(t, status)
	// Status confirmed quando receipt.status = 0x1
	assert.Equal(t, blockchain.TxStatusConfirmed, status.Status)
	assert.NotNil(t, status.BlockNumber)
	//assert.NotNil(t, status.GasUsed) // TransactionStatus não tem GasUsed
	//assert.True(t, *status.GasUsed > 0) // TransactionStatus não tem GasUsed
}

func TestEthereumProvider_GetTransactionHistory(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	history, err := provider.GetTransactionHistory(
		context.Background(),
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		&blockchain.PaginationOptions{
			Limit:  10,
			Offset: 0,
		},
	)

	assert.NoError(t, err)
	assert.NotNil(t, history)
}

func TestEthereumProvider_GetCapabilities(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	caps := provider.GetCapabilities()
	assert.True(t, caps.SupportsSmartContracts)
	assert.True(t, caps.SupportsTokens)
	assert.True(t, caps.SupportsMultiSig)
	assert.True(t, caps.RequiresGas)
}

func TestEthereumProvider_Concurrency(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	wallets := make(chan *blockchain.Wallet, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			wallet, err := provider.GenerateWallet(context.Background())
			if err != nil {
				errors <- err
				return
			}
			wallets <- wallet
		}()
	}

	generatedWallets := make([]*blockchain.Wallet, 0, 10)
	for i := 0; i < 10; i++ {
		select {
		case wallet := <-wallets:
			generatedWallets = append(generatedWallets, wallet)
		case err := <-errors:
			t.Fatal(err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for wallets")
		}
	}

	addresses := make(map[string]bool)
	for _, w := range generatedWallets {
		assert.False(t, addresses[w.Address], "duplicate address")
		addresses[w.Address] = true
	}
}

func TestEthereumProvider_ContextCancellation(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	mockRPC.handler = func(method string, params []interface{}) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return mockRPC.defaultHandler(method, params)
	}

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = provider.GetBalance(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1")
	assert.Error(t, err)
}

func TestEthereumProvider_ErrorHandling(t *testing.T) {
	mockRPC := NewMockRPCServer()
	defer mockRPC.Close()

	mockRPC.handler = func(method string, params []interface{}) (interface{}, error) {
		return nil, fmt.Errorf("RPC error")
	}

	provider, err := ethereum.NewProvider(mockRPC.URL(), blockchain.NetworkMainnet, 1)
	require.NoError(t, err)

	_, err = provider.GetBalance(context.Background(), "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1")
	assert.Error(t, err)
}
