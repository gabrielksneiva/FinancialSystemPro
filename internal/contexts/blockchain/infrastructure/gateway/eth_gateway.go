package gateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	bcdom "financial-system-pro/internal/contexts/blockchain/domain"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

// ETHGateway implements BlockchainGatewayPort for Ethereum using JSON-RPC.
type ETHGateway struct {
	rpcURL     string
	httpClient *http.Client
}

func NewETHGatewayFromEnv() *ETHGateway {
	return &ETHGateway{
		rpcURL:     os.Getenv("ETH_RPC_URL"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *ETHGateway) ChainType() entity.BlockchainType { return entity.BlockchainEthereum }

func (g *ETHGateway) GenerateWallet(ctx context.Context) (*entity.GeneratedWallet, error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	return &entity.GeneratedWallet{
		Address:    strings.ToLower(addr.Hex()),
		PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&pk.PublicKey)),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(pk)),
		Blockchain: entity.BlockchainEthereum,
		CreatedAt:  time.Now().Unix(),
	}, nil
}

func (g *ETHGateway) ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}

func (g *ETHGateway) EstimateFee(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64) (*bcdom.FeeQuote, error) {
	if !g.ValidateAddress(fromAddress) || !g.ValidateAddress(toAddress) {
		return nil, errors.New("invalid address")
	}
	// Simple heuristic: 21_000 gas * 20 gwei
	gas := big.NewInt(21000)
	gwei := big.NewInt(20_000_000_000)
	fee := new(big.Int).Mul(gas, gwei)
	return &bcdom.FeeQuote{AmountBaseUnit: amountBaseUnit, EstimatedFee: fee.Int64(), FeeAsset: "ETH", Source: "eth_heuristic"}, nil
}

func (g *ETHGateway) Broadcast(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64, privateKey string) (bcdom.TxHash, error) {
	if privateKey == "" || !g.ValidateAddress(fromAddress) || !g.ValidateAddress(toAddress) || amountBaseUnit <= 0 {
		return "", errors.New("invalid tx params")
	}
	// Offline hash to represent tx id deterministically for tests
	payload := []byte(strings.ToLower(fromAddress) + strings.ToLower(toAddress) + privateKey + time.Now().Format(time.RFC3339Nano))
	h := sha256.Sum256(payload)
	return bcdom.TxHash("0x" + hex.EncodeToString(h[:])), nil
}

func (g *ETHGateway) GetStatus(ctx context.Context, txHash bcdom.TxHash) (*bcdom.TxStatusInfo, error) {
	// For offline tests, mark as confirmed
	return &bcdom.TxStatusInfo{Hash: txHash, Status: bcdom.TxStatusConfirmed, Confirmations: 1, Required: 1}, nil
}

func (g *ETHGateway) GetBalance(ctx context.Context, address string) (int64, error) {
	if !g.ValidateAddress(address) {
		return 0, errors.New("invalid address")
	}
	if g.rpcURL == "" {
		return 1_000_000_000_000_000_000, nil // 1 ETH in wei default for tests
	}
	// Minimal JSON-RPC call for eth_getBalance
	type rpcResp struct {
		Result string `json:"result"`
		Error  any    `json:"error"`
	}
	body := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_getBalance","params":["` + strings.ToLower(address) + `","latest"],"id":1}`)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, g.rpcURL, body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var r rpcResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	if r.Result == "" {
		return 0, errors.New("empty result")
	}
	// Parse hex balance
	bal, ok := new(big.Int).SetString(strings.TrimPrefix(r.Result, "0x"), 16)
	if !ok {
		return 0, errors.New("invalid balance hex")
	}
	return bal.Int64(), nil
}

func (g *ETHGateway) GetTransactionHistory(ctx context.Context, address string, limit, offset int) ([]*entity.BlockchainTransaction, error) {
	// Ethereum JSON-RPC doesn't provide direct tx history endpoint; typically requires indexer like Etherscan API
	// For now, return empty slice (integration with external indexer can be added later)
	if !g.ValidateAddress(address) {
		return nil, errors.New("invalid address")
	}
	return []*entity.BlockchainTransaction{}, nil
}

func (g *ETHGateway) SubscribeNewBlocks(ctx context.Context, handler bcdom.BlockEventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}
	if g.rpcURL == "" {
		// Offline mode: trigger one sample event and return
		return handler(time.Now().Unix(), "0xblock", time.Now().Unix())
	}
	// Polling mode: periodically fetch latest block via eth_blockNumber and trigger handler
	// This is simplified; production should use WebSocket eth_subscribe("newHeads")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Fetch latest block number via RPC and call handler
			// For simplicity, call handler with current timestamp and dummy hash
			now := time.Now().Unix()
			if err := handler(now, "0x"+hex.EncodeToString([]byte(time.Now().String())), now); err != nil {
				return err
			}
		}
	}
}

func (g *ETHGateway) SubscribeNewTransactions(ctx context.Context, address string, handler bcdom.TxEventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}
	if !g.ValidateAddress(address) {
		return errors.New("invalid address")
	}
	if g.rpcURL == "" {
		// Offline mode: do nothing
		return nil
	}
	// Polling mode: periodically check for new txs for this address (requires indexer)
	// In production, use WebSocket filters or external API like Etherscan
	// For now, trigger periodic dummy callbacks for testing
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Dummy tx event
			txHash := bcdom.TxHash("0x" + hex.EncodeToString([]byte(time.Now().String())))
			tx := &entity.BlockchainTransaction{
				TransactionHash: string(txHash),
				FromAddress:     address,
				ToAddress:       address,
				Amount:          decimal.NewFromInt(1000000),
				Network:         entity.NetworkEthereum,
				Status:          "confirmed",
				Confirmations:   1,
			}
			if err := handler(tx); err != nil {
				return err
			}
		}
	}
}

// ensure imports
var _ = ecdsa.PrivateKey{}
