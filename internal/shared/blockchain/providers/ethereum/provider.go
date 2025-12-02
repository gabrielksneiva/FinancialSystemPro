package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"financial-system-pro/internal/shared/blockchain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

// Provider implementa blockchain.Provider para Ethereum
type Provider struct {
	rpcURL      string
	networkType blockchain.NetworkType
	chainID     *big.Int
	client      *RPCClient
}

// NewProvider cria um novo provider Ethereum
func NewProvider(rpcURL string, networkType blockchain.NetworkType, chainID int64) (*Provider, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("rpcURL cannot be empty")
	}

	return &Provider{
		rpcURL:      rpcURL,
		networkType: networkType,
		chainID:     big.NewInt(chainID),
		client:      NewRPCClient(rpcURL),
	}, nil
}

// ChainType retorna o tipo de blockchain
func (p *Provider) ChainType() blockchain.ChainType {
	return blockchain.ChainEthereum
}

// NetworkType retorna o tipo de rede
func (p *Provider) NetworkType() blockchain.NetworkType {
	return p.networkType
}

// IsHealthy verifica se o provider está saudável
func (p *Provider) IsHealthy(ctx context.Context) bool {
	_, err := p.client.Call(ctx, "eth_blockNumber", nil)
	return err == nil
}

// GenerateWallet gera uma nova carteira Ethereum
func (p *Provider) GenerateWallet(ctx context.Context) (*blockchain.Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GenerateWallet", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GenerateWallet", fmt.Errorf("failed to cast public key"))
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	privateKeyBytes := crypto.FromECDSA(privateKey)

	return &blockchain.Wallet{
		Address:    address.Hex(),
		PublicKey:  hexutil.Encode(crypto.FromECDSAPub(publicKeyECDSA)),
		PrivateKey: hexutil.Encode(privateKeyBytes),
		ChainType:  blockchain.ChainEthereum,
		CreatedAt:  time.Now(),
	}, nil
}

// ValidateAddress valida um endereço Ethereum
func (p *Provider) ValidateAddress(address string) error {
	if !common.IsHexAddress(address) {
		return blockchain.NewValidationError("address", "invalid ethereum address format")
	}
	return nil
}

// ImportWallet importa uma carteira a partir de uma chave privada
func (p *Provider) ImportWallet(ctx context.Context, privateKeyHex string) (*blockchain.Wallet, error) {
	// Remove prefixo 0x se existir
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "ImportWallet", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "ImportWallet", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "ImportWallet", fmt.Errorf("failed to cast public key"))
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &blockchain.Wallet{
		Address:    address.Hex(),
		PublicKey:  hexutil.Encode(crypto.FromECDSAPub(publicKeyECDSA)),
		PrivateKey: "0x" + privateKeyHex,
		ChainType:  blockchain.ChainEthereum,
		CreatedAt:  time.Now(),
	}, nil
}

// GetBalance retorna o saldo de um endereço
func (p *Provider) GetBalance(ctx context.Context, address string) (*blockchain.Balance, error) {
	if err := p.ValidateAddress(address); err != nil {
		return nil, err
	}

	params := []interface{}{address, "latest"}
	result, err := p.client.Call(ctx, "eth_getBalance", params)
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GetBalance", err)
	}

	var balanceHex string
	if err := json.Unmarshal(result, &balanceHex); err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GetBalance", err)
	}

	balanceWei, ok := new(big.Int).SetString(strings.TrimPrefix(balanceHex, "0x"), 16)
	if !ok {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GetBalance", fmt.Errorf("invalid balance format"))
	}

	// Converter Wei para ETH (18 decimais)
	balanceETH := new(big.Float).Quo(
		new(big.Float).SetInt(balanceWei),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	balanceDecimal, _ := decimal.NewFromString(balanceETH.String())

	// Obter block number atual
	blockNumResult, err := p.client.Call(ctx, "eth_blockNumber", nil)
	var blockNumber int64
	if err == nil {
		var blockHex string
		if err := json.Unmarshal(blockNumResult, &blockHex); err == nil {
			if blockInt, ok := new(big.Int).SetString(strings.TrimPrefix(blockHex, "0x"), 16); ok {
				blockNumber = blockInt.Int64()
			}
		}
	}

	return &blockchain.Balance{
		Address:       address,
		Amount:        balanceDecimal,
		AmountRaw:     balanceWei.String(),
		Currency:      "ETH",
		Decimals:      18,
		BlockNumber:   blockNumber,
		LastUpdatedAt: time.Now(),
	}, nil
}

// EstimateFee estima a taxa de transação
func (p *Provider) EstimateFee(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.FeeEstimate, error) {
	// Obter preço de gas atual
	result, err := p.client.Call(ctx, "eth_gasPrice", nil)
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "EstimateFee", err)
	}

	var gasPriceHex string
	if err := json.Unmarshal(result, &gasPriceHex); err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "EstimateFee", err)
	}

	gasPrice, ok := new(big.Int).SetString(strings.TrimPrefix(gasPriceHex, "0x"), 16)
	if !ok {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "EstimateFee", fmt.Errorf("invalid gas price"))
	}

	// Estimar gas limit
	var gasLimit uint64 = 21000 // Transferência simples
	if len(intent.Data) > 0 {
		// Para smart contracts, seria necessário estimar
		gasLimit = 100000
	}

	// Calcular custo total em Wei
	totalCostWei := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))

	// Converter para ETH
	totalCostETH := new(big.Float).Quo(
		new(big.Float).SetInt(totalCostWei),
		new(big.Float).SetInt(big.NewInt(1e18)),
	)

	baseFee, _ := decimal.NewFromString(totalCostETH.String())

	// Calcular estimativas low, medium, high (variações de 20%)
	low := baseFee.Mul(decimal.NewFromFloat(0.8))
	medium := baseFee
	high := baseFee.Mul(decimal.NewFromFloat(1.5))

	gasPriceStr := gasPriceHex

	return &blockchain.FeeEstimate{
		ChainType:   blockchain.ChainEthereum,
		Low:         low,
		Medium:      medium,
		High:        high,
		Currency:    "ETH",
		GasPrice:    &gasPriceStr,
		GasLimit:    &gasLimit,
		EstimatedAt: time.Now(),
	}, nil
}

// BuildTransaction constrói uma transação não assinada
func (p *Provider) BuildTransaction(ctx context.Context, intent *blockchain.TransactionIntent) (*blockchain.UnsignedTransaction, error) {
	if err := p.ValidateAddress(intent.From); err != nil {
		return nil, err
	}
	if err := p.ValidateAddress(intent.To); err != nil {
		return nil, err
	}

	// Obter nonce
	var nonce uint64
	if intent.Nonce != nil {
		nonce = *intent.Nonce
	} else {
		params := []interface{}{intent.From, "latest"}
		result, err := p.client.Call(ctx, "eth_getTransactionCount", params)
		if err != nil {
			return nil, blockchain.NewChainError(blockchain.ChainEthereum, "BuildTransaction", err)
		}

		var nonceHex string
		if err := json.Unmarshal(result, &nonceHex); err != nil {
			return nil, blockchain.NewChainError(blockchain.ChainEthereum, "BuildTransaction", err)
		}

		nonceBig, ok := new(big.Int).SetString(strings.TrimPrefix(nonceHex, "0x"), 16)
		if !ok {
			return nil, blockchain.NewChainError(blockchain.ChainEthereum, "BuildTransaction", fmt.Errorf("invalid nonce"))
		}
		nonce = nonceBig.Uint64()
	}

	// Estimar fee
	feeEstimate, err := p.EstimateFee(ctx, intent)
	if err != nil {
		return nil, err
	}

	rawData := map[string]interface{}{
		"from":     intent.From,
		"to":       intent.To,
		"value":    intent.AmountRaw,
		"nonce":    nonce,
		"gasPrice": *feeEstimate.GasPrice,
		"gasLimit": *feeEstimate.GasLimit,
		"chainId":  p.chainID.String(),
	}

	if len(intent.Data) > 0 {
		rawData["data"] = hexutil.Encode(intent.Data)
	}

	return &blockchain.UnsignedTransaction{
		ChainType: blockchain.ChainEthereum,
		From:      intent.From,
		To:        intent.To,
		Amount:    intent.Amount,
		Fee:       feeEstimate.Medium,
		Nonce:     nonce,
		Data:      intent.Data,
		RawData:   rawData,
		CreatedAt: time.Now(),
	}, nil
}

// SignTransaction assina uma transação
func (p *Provider) SignTransaction(ctx context.Context, tx *blockchain.UnsignedTransaction, privateKey *blockchain.PrivateKey) (*blockchain.SignedTransaction, error) {
	// Implementação simplificada - na produção usar biblioteca completa
	return &blockchain.SignedTransaction{
		ChainType: blockchain.ChainEthereum,
		RawTx:     "0x" + hex.EncodeToString(privateKey.Raw),
		TxHash:    "0x" + hex.EncodeToString(crypto.Keccak256(privateKey.Raw)),
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Signature: "0xsignature",
		RawData:   tx.RawData,
		SignedAt:  time.Now(),
	}, nil
}

// BroadcastTransaction transmite uma transação assinada
func (p *Provider) BroadcastTransaction(ctx context.Context, tx *blockchain.SignedTransaction) (*blockchain.TransactionReceipt, error) {
	params := []interface{}{tx.RawTx}
	result, err := p.client.Call(ctx, "eth_sendRawTransaction", params)
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "BroadcastTransaction", err)
	}

	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "BroadcastTransaction", err)
	}

	return &blockchain.TransactionReceipt{
		TxHash:        txHash,
		ChainType:     blockchain.ChainEthereum,
		Status:        blockchain.TxStatusPending,
		From:          tx.From,
		To:            tx.To,
		Amount:        tx.Amount,
		Fee:           tx.Fee,
		Confirmations: 0,
		BroadcastAt:   time.Now(),
	}, nil
}

// GetTransactionStatus retorna o status de uma transação
func (p *Provider) GetTransactionStatus(ctx context.Context, txHash string) (*blockchain.TransactionStatus, error) {
	params := []interface{}{txHash}
	result, err := p.client.Call(ctx, "eth_getTransactionReceipt", params)
	if err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GetTransactionStatus", err)
	}

	var receipt map[string]interface{}
	if err := json.Unmarshal(result, &receipt); err != nil {
		return nil, blockchain.NewChainError(blockchain.ChainEthereum, "GetTransactionStatus", err)
	}

	if receipt == nil {
		return &blockchain.TransactionStatus{
			TxHash: txHash,
			Status: blockchain.TxStatusPending,
		}, nil
	}

	status := blockchain.TxStatusConfirmed
	if statusHex, ok := receipt["status"].(string); ok {
		if statusHex == "0x0" {
			status = blockchain.TxStatusFailed
		}
	}

	var blockNumber *int64
	if blockNumHex, ok := receipt["blockNumber"].(string); ok {
		if blockInt, ok := new(big.Int).SetString(strings.TrimPrefix(blockNumHex, "0x"), 16); ok {
			bn := blockInt.Int64()
			blockNumber = &bn
		}
	}

	return &blockchain.TransactionStatus{
		TxHash:      txHash,
		Status:      status,
		BlockNumber: blockNumber,
	}, nil
}

// GetTransactionHistory retorna o histórico de transações (simplificado)
func (p *Provider) GetTransactionHistory(ctx context.Context, address string, opts *blockchain.PaginationOptions) (*blockchain.TransactionHistory, error) {
	// Esta implementação requer um serviço de indexação como Etherscan API
	// Por enquanto, retorna vazio
	return &blockchain.TransactionHistory{
		Address:      address,
		Transactions: []blockchain.HistoricalTransaction{},
		Total:        0,
		HasMore:      false,
	}, nil
}

// SubscribeNewBlocks subscreve a novos blocos (não implementado)
func (p *Provider) SubscribeNewBlocks(ctx context.Context, handler blockchain.BlockHandler) error {
	return blockchain.ErrNotSupported
}

// SubscribeNewTransactions subscreve a novas transações (não implementado)
func (p *Provider) SubscribeNewTransactions(ctx context.Context, filter *blockchain.TxFilter, handler blockchain.TxHandler) error {
	return blockchain.ErrNotSupported
}

// UnsubscribeAll cancela todas as subscrições
func (p *Provider) UnsubscribeAll(ctx context.Context) error {
	return nil
}

// GetCapabilities retorna as capabilities do provider
func (p *Provider) GetCapabilities() *blockchain.ProviderCapabilities {
	return &blockchain.ProviderCapabilities{
		SupportsSmartContracts:   true,
		SupportsTokens:           true,
		SupportsStaking:          true,
		SupportsSubscriptions:    false, // Requer WebSocket
		SupportsMemo:             false,
		SupportsMultiSig:         true,
		RequiresGas:              true,
		NativeTokenDecimals:      18,
		MinConfirmationsRequired: 12,
		AverageBlockTime:         12 * time.Second,
	}
}
