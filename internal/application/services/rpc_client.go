package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// JSONRPCRequest representa uma requisição JSON-RPC 2.0
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int64         `json:"id"`
}

// JSONRPCResponse representa uma resposta JSON-RPC 2.0
type JSONRPCResponse struct {
	Error   *JSONRPCError   `json:"error"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	ID      int64           `json:"id"`
}

// JSONRPCError representa um erro JSON-RPC
type JSONRPCError struct {
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
	Code    int    `json:"code"`
}

// RPCClient representa um cliente RPC com pool de conexões
type RPCClient struct {
	httpClient *http.Client
	endpoint   string
	requestID  int64
	mu         sync.Mutex
}

// NewRPCClient cria um novo cliente RPC otimizado
func NewRPCClient(endpoint string) *RPCClient {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &RPCClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		},
		requestID: 1,
	}
}

// Call executa uma chamada RPC
func (c *RPCClient) Call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	id := c.requestID
	c.requestID++
	c.mu.Unlock()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao marshalar requisição: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "FinancialSystemPro/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("erro RPC %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// GetBalance obtém o saldo usando JSON-RPC
func (c *RPCClient) GetBalance(ctx context.Context, address string) (int64, error) {
	result, err := c.Call(ctx, "eth_getBalance", address, "latest")
	if err != nil {
		return 0, err
	}

	var balance string
	if err := json.Unmarshal(result, &balance); err != nil {
		return 0, fmt.Errorf("erro ao decodificar saldo: %w", err)
	}

	// Converter de hex para int64
	var balanceInt int64
	if _, err := fmt.Sscanf(balance, "%x", &balanceInt); err != nil {
		return 0, fmt.Errorf("erro ao converter saldo de hex: %w", err)
	}
	return balanceInt, nil
}

// GetBlockNumber obtém o número do bloco atual
func (c *RPCClient) GetBlockNumber(ctx context.Context) (int64, error) {
	result, err := c.Call(ctx, "eth_blockNumber")
	if err != nil {
		return 0, err
	}

	var blockNum string
	if err := json.Unmarshal(result, &blockNum); err != nil {
		return 0, fmt.Errorf("erro ao decodificar número do bloco: %w", err)
	}

	var blockNumInt int64
	if _, err := fmt.Sscanf(blockNum, "%x", &blockNumInt); err != nil {
		return 0, fmt.Errorf("erro ao converter número do bloco de hex: %w", err)
	}
	return blockNumInt, nil
}

// GetTransactionStatus obtém o status de uma transação
func (c *RPCClient) GetTransactionStatus(ctx context.Context, txHash string) (map[string]interface{}, error) {
	result, err := c.Call(ctx, "eth_getTransactionReceipt", txHash)
	if err != nil {
		return nil, err
	}

	var txReceipt map[string]interface{}
	if err := json.Unmarshal(result, &txReceipt); err != nil {
		return nil, fmt.Errorf("erro ao decodificar recibo da transação: %w", err)
	}

	return txReceipt, nil
}

// EstimateGas estima o gasto de gas/energia
func (c *RPCClient) EstimateGas(ctx context.Context, from, to string, value, data string) (int64, error) {
	txObj := map[string]string{
		"from":  from,
		"to":    to,
		"value": value,
	}

	if data != "" {
		txObj["data"] = data
	}

	result, err := c.Call(ctx, "eth_estimateGas", txObj)
	if err != nil {
		return 0, err
	}

	var gasEstimate string
	if err := json.Unmarshal(result, &gasEstimate); err != nil {
		return 0, fmt.Errorf("erro ao decodificar estimativa de gas: %w", err)
	}

	var gasInt int64
	if _, err := fmt.Sscanf(gasEstimate, "%x", &gasInt); err != nil {
		return 0, fmt.Errorf("erro ao converter estimativa de gas de hex: %w", err)
	}
	return gasInt, nil
}

// SendRawTransaction envia uma transação assinada
func (c *RPCClient) SendRawTransaction(ctx context.Context, signedTx string) (string, error) {
	result, err := c.Call(ctx, "eth_sendRawTransaction", signedTx)
	if err != nil {
		return "", err
	}

	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return "", fmt.Errorf("erro ao decodificar hash da transação: %w", err)
	}

	return txHash, nil
}

// GetTransaction obtém informações de uma transação
func (c *RPCClient) GetTransaction(ctx context.Context, txHash string) (map[string]interface{}, error) {
	result, err := c.Call(ctx, "eth_getTransactionByHash", txHash)
	if err != nil {
		return nil, err
	}

	var tx map[string]interface{}
	if err := json.Unmarshal(result, &tx); err != nil {
		return nil, fmt.Errorf("erro ao decodificar transação: %w", err)
	}

	return tx, nil
}

// GetCode obtém o bytecode de um contrato
func (c *RPCClient) GetCode(ctx context.Context, address string) (string, error) {
	result, err := c.Call(ctx, "eth_getCode", address, "latest")
	if err != nil {
		return "", err
	}

	var code string
	if err := json.Unmarshal(result, &code); err != nil {
		return "", fmt.Errorf("erro ao decodificar código: %w", err)
	}

	return code, nil
}

// Call é um alias para chamar um contrato (view/pure)
func (c *RPCClient) CallContract(ctx context.Context, from, to, data string) (string, error) {
	txObj := map[string]string{
		"from": from,
		"to":   to,
		"data": data,
	}

	result, err := c.Call(ctx, "eth_call", txObj, "latest")
	if err != nil {
		return "", err
	}

	var callResult string
	if err := json.Unmarshal(result, &callResult); err != nil {
		return "", fmt.Errorf("erro ao decodificar resultado da chamada: %w", err)
	}

	return callResult, nil
}

// GetGasPrice obtém o preço do gas atual
func (c *RPCClient) GetGasPrice(ctx context.Context) (int64, error) {
	result, err := c.Call(ctx, "eth_gasPrice")
	if err != nil {
		return 0, err
	}

	var gasPriceStr string
	if err := json.Unmarshal(result, &gasPriceStr); err != nil {
		return 0, fmt.Errorf("erro ao decodificar preço do gas: %w", err)
	}

	var gasPrice int64
	if _, err := fmt.Sscanf(gasPriceStr, "%x", &gasPrice); err != nil {
		return 0, fmt.Errorf("erro ao converter preço do gas de hex: %w", err)
	}
	return gasPrice, nil
}

// Close fecha o cliente RPC
func (c *RPCClient) Close() {
	c.httpClient.CloseIdleConnections()
}
