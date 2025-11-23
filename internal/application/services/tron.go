package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"financial-system-pro/internal/domain/entities"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
)

type TronService struct {
	testnetRPC      string
	testnetGRPC     string
	apiKey          string
	vaultAddress    string // Endereço do cofre TRON
	vaultPrivateKey string // Private key do cofre
	httpClient      *http.Client
	rpcClient       *RPCClient
	grpcClient      *TronGRPCClient
	mu              sync.RWMutex
	lastRPCError    error
	lastRPCErrorAt  time.Time
}

// NewTronService inicializa a conexão com Tron Testnet
// Agora recebe config para acessar credenciais do cofre
func NewTronService(vaultAddress, vaultPrivateKey string) *TronService {
	rpcEndpoint := os.Getenv("TRON_TESTNET_RPC")
	grpcEndpoint := os.Getenv("TRON_TESTNET_GRPC")

	ts := &TronService{
		testnetRPC:      rpcEndpoint,
		testnetGRPC:     grpcEndpoint,
		apiKey:          os.Getenv("TRON_API_KEY"),
		vaultAddress:    vaultAddress,
		vaultPrivateKey: vaultPrivateKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Inicializar cliente RPC JSON-RPC
	if rpcEndpoint != "" {
		ts.rpcClient = NewRPCClient(rpcEndpoint)
	}

	// Inicializar cliente gRPC (opcional, para performance)
	if grpcEndpoint != "" {
		if grpcCli, err := NewTronGRPCClient(grpcEndpoint); err == nil {
			ts.grpcClient = grpcCli
		}
	}

	return ts
}

// GetBalance retorna o saldo de uma conta Tron em SUN (1 TRX = 1.000.000 SUN)
func (ts *TronService) GetBalance(address string) (int64, error) {
	if address == "" {
		return 0, fmt.Errorf("endereço inválido")
	}

	if !ts.ValidateAddress(address) {
		return 0, fmt.Errorf("endereço Tron inválido")
	}

	// Fazer requisição à API Tron
	url := fmt.Sprintf("%s/v1/accounts/%s", ts.testnetRPC, address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("erro na API Tron: status %d - %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Address string `json:"address"`
		Balance int64  `json:"balance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return apiResp.Balance, nil
}

// ValidateAddress valida um endereço Tron
func (ts *TronService) ValidateAddress(address string) bool {
	if len(address) != 34 {
		return false
	}
	if address[0] != 'T' {
		return false
	}
	return true
}

// Base58ToHex converte um endereço TRON de Base58 para Hexadecimal
// A TRON API espera endereços em formato hex, não base58
func (ts *TronService) Base58ToHex(address string) (string, error) {
	// Decodificar Base58
	decoded := base58.Decode(address)
	if len(decoded) == 0 {
		return "", fmt.Errorf("endereço base58 inválido")
	}

	// Remover os últimos 4 bytes (checksum)
	if len(decoded) < 4 {
		return "", fmt.Errorf("endereço muito curto")
	}
	addressBytes := decoded[:len(decoded)-4]

	// Converter para hexadecimal
	return hex.EncodeToString(addressBytes), nil
}

// CreateWallet cria uma nova carteira Tron
func (ts *TronService) CreateWallet() (*entities.TronWallet, error) {
	// Usar a API Tron para gerar uma carteira
	url := fmt.Sprintf("%s/wallet/createaccount", ts.testnetRPC)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na API Tron: status %d - %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Address    string `json:"address"`
		Pubkey     string `json:"pubkey"`
		Prikey     string `json:"prikey"`
		PrivateKey string `json:"private_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	privateKey := apiResp.PrivateKey
	if privateKey == "" {
		privateKey = apiResp.Prikey
	}

	wallet := &entities.TronWallet{
		Address:    apiResp.Address,
		PrivateKey: privateKey,
		PublicKey:  apiResp.Pubkey,
	}

	return wallet, nil
}

// SendTransaction envia uma transação na rede Tron
func (ts *TronService) SendTransaction(fromAddress, toAddress string, amount int64, privateKey string) (string, error) {
	if !ts.ValidateAddress(fromAddress) {
		return "", fmt.Errorf("endereço de origem inválido")
	}
	if !ts.ValidateAddress(toAddress) {
		return "", fmt.Errorf("endereço de destino inválido")
	}
	if amount <= 0 {
		return "", fmt.Errorf("valor deve ser maior que zero")
	}
	if privateKey == "" {
		return "", fmt.Errorf("chave privada é obrigatória")
	}

	// Converter endereços Base58 para Hexadecimal
	fromAddressHex, err := ts.Base58ToHex(fromAddress)
	if err != nil {
		return "", fmt.Errorf("erro ao converter endereço de origem: %w", err)
	}

	toAddressHex, err := ts.Base58ToHex(toAddress)
	if err != nil {
		return "", fmt.Errorf("erro ao converter endereço de destino: %w", err)
	}

	// Criar transação
	url := fmt.Sprintf("%s/wallet/createtransaction", ts.testnetRPC)

	payload := map[string]interface{}{
		"to_address":    toAddressHex,
		"owner_address": fromAddressHex,
		"amount":        amount,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	// Ler o body completo para debug
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro na API Tron: status %d - %s", resp.StatusCode, string(bodyBytes))
	}

	// Log da resposta completa para debug
	fmt.Printf("🔍 TRON API Response: %s\n", string(bodyBytes))

	var createTxResp struct {
		TxID       string                 `json:"txID"`
		RawDataHex string                 `json:"raw_data_hex"`
		RawData    map[string]interface{} `json:"raw_data"`
		Error      string                 `json:"Error"`
	}

	if err := json.Unmarshal(bodyBytes, &createTxResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %w - body: %s", err, string(bodyBytes))
	}

	// Verificar se há erro na resposta
	if createTxResp.Error != "" {
		return "", fmt.Errorf("erro da API TRON: %s", createTxResp.Error)
	}

	txID := createTxResp.TxID
	if txID == "" {
		return "", fmt.Errorf("txID não retornado pela API - response: %s", string(bodyBytes))
	}

	// Passo 2: Assinar a transação
	signedTx, err := ts.signTransaction(bodyBytes, privateKey)
	if err != nil {
		return "", fmt.Errorf("erro ao assinar transação: %w", err)
	}

	// Passo 3: Transmitir a transação assinada
	broadcastErr := ts.broadcastTransaction(signedTx)
	if broadcastErr != nil {
		return "", fmt.Errorf("erro ao transmitir transação: %w", broadcastErr)
	}

	return txID, nil
}

// GetTransactionStatus obtém o status de uma transação
func (ts *TronService) GetTransactionStatus(txHash string) (string, error) {
	if txHash == "" {
		return "", fmt.Errorf("hash da transação inválido")
	}

	// Obter informações da transação
	url := fmt.Sprintf("%s/walletsolidity/gettransactionbyid", ts.testnetRPC)

	payload := map[string]interface{}{
		"value": txHash,
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "not_found", nil
	}

	var apiResp struct {
		TxID        string `json:"txID"`
		BlockNumber int64  `json:"block_number"`
		Confirmed   bool   `json:"confirmed"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "unknown", nil
	}

	_ = payloadBytes

	if apiResp.Confirmed {
		return "confirmed", nil
	}

	if apiResp.BlockNumber > 0 {
		return "in_progress", nil
	}

	return "pending", nil
}

// GetTransaction obtém informações detalhadas de uma transação (implementa TronAPI interface)
func (ts *TronService) GetTransaction(txHash string) (interface{}, error) {
	if txHash == "" {
		return nil, fmt.Errorf("hash da transação inválido")
	}

	// Obter informações da transação
	url := fmt.Sprintf("%s/walletsolidity/gettransactionbyid", ts.testnetRPC)

	payload := map[string]interface{}{
		"value": txHash,
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewBuffer(payloadBytes))

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transação não encontrada: status %d", resp.StatusCode)
	}

	var transaction map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		return nil, fmt.Errorf("erro ao decodificar transação: %w", err)
	}

	return transaction, nil
}

// EstimateGasForTransaction estima o gasto de energia para uma transação
func (ts *TronService) EstimateGasForTransaction(fromAddress, toAddress string, amount int64) (int64, error) {
	if !ts.ValidateAddress(fromAddress) || !ts.ValidateAddress(toAddress) {
		return 0, fmt.Errorf("endereços inválidos")
	}

	if amount <= 0 {
		return 0, fmt.Errorf("valor deve ser maior que zero")
	}

	// Consultar informações sobre energia no Tron
	url := fmt.Sprintf("%s/wallet/getaccount", ts.testnetRPC)

	payload := map[string]interface{}{
		"address": fromAddress,
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("erro na API Tron: status %d - %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		EnergyUsage      int64                  `json:"energy_usage"`
		EnergyUsageTotal int64                  `json:"energy_usage_total"`
		AccountResource  map[string]interface{} `json:"account_resource"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		// Estimativa padrão em caso de erro
		return 25000, nil
	}

	_ = payloadBytes

	// Estimativa: transação simples de transferência usa aproximadamente 25000 energia
	estimatedGas := int64(25000)

	if apiResp.EnergyUsage > 0 {
		estimatedGas = apiResp.EnergyUsage
	}

	return estimatedGas, nil
}

// ConvertSunToTRX converte SUN para TRX (1 TRX = 1,000,000 SUN)
func (ts *TronService) ConvertSunToTRX(sun int64) float64 {
	return float64(sun) / 1_000_000
}

// ConvertTRXToSun converte TRX para SUN (1 TRX = 1,000,000 SUN)
func (ts *TronService) ConvertTRXToSun(trx float64) int64 {
	return int64(trx * 1_000_000)
}

// IsTestnetConnected verifica se está conectado à testnet Tron
func (ts *TronService) IsTestnetConnected() bool {
	if ts.testnetRPC == "" {
		return false
	}

	url := fmt.Sprintf("%s/wallet/getnowblock", ts.testnetRPC)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetNetworkInfo obtém informações sobre a rede Tron
func (ts *TronService) GetNetworkInfo() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/wallet/getnowblock", ts.testnetRPC)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição à API Tron: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API Tron: status %d", resp.StatusCode)
	}

	var blockInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&blockInfo); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return blockInfo, nil
}

// GetRPCClient retorna o cliente RPC com verificações de saúde
func (ts *TronService) GetRPCClient() *RPCClient {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.rpcClient
}

// GetGRPCClient retorna o cliente gRPC com verificações de saúde
func (ts *TronService) GetGRPCClient() *TronGRPCClient {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.grpcClient
}

// Close fecha todos os clientes
func (ts *TronService) Close() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.rpcClient != nil {
		ts.rpcClient.Close()
	}

	if ts.grpcClient != nil {
		return ts.grpcClient.Close()
	}

	return nil
}

// HasGRPCSupport verifica se gRPC está disponível
func (ts *TronService) HasGRPCSupport() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if ts.grpcClient == nil {
		return false
	}

	return ts.grpcClient.IsConnected()
}

// GetConnectionStatus retorna o status das conexões
func (ts *TronService) GetConnectionStatus() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	status := make(map[string]interface{})

	status["rpc_enabled"] = ts.rpcClient != nil
	status["grpc_enabled"] = ts.grpcClient != nil

	if ts.grpcClient != nil {
		status["grpc_connected"] = ts.grpcClient.IsConnected()
	}

	status["last_error"] = ts.lastRPCError
	status["last_error_time"] = ts.lastRPCErrorAt

	return status
}

// RecordError registra um erro para monitoramento
func (ts *TronService) RecordError(err error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.lastRPCError = err
	ts.lastRPCErrorAt = time.Now()
}

// HealthCheck verifica se Tron RPC está respondendo
func (ts *TronService) HealthCheck(ctx context.Context) error {
	if ts.testnetRPC == "" {
		return fmt.Errorf("tron rpc endpoint not configured")
	}

	// Criar um request leve para verificar conexão
	req, err := http.NewRequestWithContext(ctx, "POST", ts.testnetRPC, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tron rpc health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("tron rpc returned status %d", resp.StatusCode)
	}

	return nil
}

// signTransaction assina uma transação TRON localmente usando ECDSA
func (ts *TronService) signTransaction(unsignedTxBytes []byte, privateKeyHex string) ([]byte, error) {
	// Parse da transação não assinada
	var txData struct {
		TxID       string `json:"txID"`
		RawDataHex string `json:"raw_data_hex"`
	}

	if err := json.Unmarshal(unsignedTxBytes, &txData); err != nil {
		return nil, fmt.Errorf("erro ao parsear transação: %w", err)
	}

	// Decodificar o raw_data_hex
	rawDataBytes, err := hex.DecodeString(txData.RawDataHex)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar raw_data_hex: %w", err)
	}

	// Calcular hash SHA-256 do raw_data
	hash := sha256.Sum256(rawDataBytes)

	// Converter private key hex para ECDSA private key
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar chave privada: %w", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter chave privada: %w", err)
	}

	// Assinar o hash
	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar transação: %w", err)
	}

	// TRON usa signature completa (65 bytes incluindo recovery ID)
	// Converter signature para hex (todos os 65 bytes)
	signatureHex := hex.EncodeToString(signature)

	// Criar estrutura da transação assinada
	var txMap map[string]interface{}
	if err := json.Unmarshal(unsignedTxBytes, &txMap); err != nil {
		return nil, fmt.Errorf("erro ao parsear transação original: %w", err)
	}

	// Adicionar assinatura
	txMap["signature"] = []string{signatureHex}

	// Serializar transação assinada
	signedTx, err := json.Marshal(txMap)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar transação assinada: %w", err)
	}

	fmt.Printf("🔏 Transação assinada: %s\n", string(signedTx))

	return signedTx, nil
}

// broadcastTransaction transmite uma transação assinada para a blockchain TRON
func (ts *TronService) broadcastTransaction(signedTxBytes []byte) error {
	url := fmt.Sprintf("%s/wallet/broadcasttransaction", ts.testnetRPC)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(signedTxBytes))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de broadcast: %w", err)
	}

	if ts.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", ts.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de broadcast: %w", err)
	}
	defer resp.Body.Close()

	broadcastBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler resposta de broadcast: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro na API de broadcast: status %d - %s", resp.StatusCode, string(broadcastBytes))
	}

	var broadcastResp struct {
		Result  bool   `json:"result"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(broadcastBytes, &broadcastResp); err != nil {
		return fmt.Errorf("erro ao decodificar resposta de broadcast: %w - body: %s", err, string(broadcastBytes))
	}

	if !broadcastResp.Result {
		return fmt.Errorf("broadcast falhou: %s (code: %s)", broadcastResp.Message, broadcastResp.Code)
	}

	fmt.Printf("📡 Transação transmitida com sucesso: %s\n", string(broadcastBytes))

	return nil
}

// GetVaultAddress retorna o endereço do cofre TRON
func (ts *TronService) GetVaultAddress() string {
	return ts.vaultAddress
}

// GetVaultPrivateKey retorna a private key do cofre TRON
func (ts *TronService) GetVaultPrivateKey() string {
	return ts.vaultPrivateKey
}

// HasVaultConfigured verifica se o cofre está configurado
func (ts *TronService) HasVaultConfigured() bool {
	return ts.vaultAddress != "" && ts.vaultPrivateKey != ""
}
