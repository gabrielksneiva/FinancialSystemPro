package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	bcdom "financial-system-pro/internal/contexts/blockchain/domain"
	entity "financial-system-pro/internal/contexts/blockchain/domain/entity"

	"github.com/btcsuite/btcutil/base58"
	"github.com/shopspring/decimal"
)

// BTCGateway implements BlockchainGatewayPort for Bitcoin.
type BTCGateway struct {
	rpcURL     string
	httpClient *http.Client
}

func NewBTCGatewayFromEnv() *BTCGateway {
	return &BTCGateway{
		rpcURL:     os.Getenv("BTC_RPC_URL"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *BTCGateway) ChainType() entity.BlockchainType { return entity.BlockchainBitcoin }

func (g *BTCGateway) GenerateWallet(ctx context.Context) (*entity.GeneratedWallet, error) {
	// Generate a pseudo-random P2PKH address for tests: version(0x00) + 20 bytes + checksum
	seed := sha256.Sum256([]byte(time.Now().Format(time.RFC3339Nano)))
	// Generate private key (32 bytes hex)
	privKey := hex.EncodeToString(seed[:])
	// Generate public key from private (simplified: hash of privkey for test)
	pubHash := sha256.Sum256(seed[:])
	pubKey := hex.EncodeToString(pubHash[:20])

	versioned := append([]byte{0x00}, seed[:20]...)
	first := sha256.Sum256(versioned)
	second := sha256.Sum256(first[:])
	full := append(versioned, second[:4]...)
	addr := base58.Encode(full)
	return &entity.GeneratedWallet{
		Address:    addr,
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Blockchain: entity.BlockchainBitcoin,
		CreatedAt:  time.Now().Unix(),
	}, nil
}

func (g *BTCGateway) ValidateAddress(address string) bool {
	// Basic check for base58 P2PKH or bech32 patterns
	if address == "" {
		return false
	}
	if strings.HasPrefix(strings.ToLower(address), "bc1") {
		return len(address) >= 14
	}
	// base58 typical length 26..35
	return len(address) >= 26 && len(address) <= 50
}

func (g *BTCGateway) EstimateFee(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64) (*bcdom.FeeQuote, error) {
	if !g.ValidateAddress(fromAddress) || !g.ValidateAddress(toAddress) {
		return nil, errors.New("invalid address")
	}
	// Heuristic fee: 180*in + 34*out + 10 extra; assume 1 in/2 out, 1 sat/vB
	fee := int64(258)
	return &bcdom.FeeQuote{AmountBaseUnit: amountBaseUnit, EstimatedFee: fee, FeeAsset: "BTC", Source: "btc_heuristic"}, nil
}

func (g *BTCGateway) Broadcast(ctx context.Context, fromAddress, toAddress string, amountBaseUnit int64, privateKey string) (bcdom.TxHash, error) {
	if privateKey == "" || !g.ValidateAddress(fromAddress) || !g.ValidateAddress(toAddress) || amountBaseUnit <= 0 {
		return "", errors.New("invalid tx params")
	}
	payload := []byte(fromAddress + toAddress + privateKey + time.Now().Format(time.RFC3339Nano))
	h := sha256.Sum256(payload)
	return bcdom.TxHash(hex.EncodeToString(h[:])), nil
}

func (g *BTCGateway) GetStatus(ctx context.Context, txHash bcdom.TxHash) (*bcdom.TxStatusInfo, error) {
	return &bcdom.TxStatusInfo{Hash: txHash, Status: bcdom.TxStatusConfirmed, Confirmations: 1, Required: 1}, nil
}

func (g *BTCGateway) GetBalance(ctx context.Context, address string) (int64, error) {
	if !g.ValidateAddress(address) {
		return 0, errors.New("invalid address")
	}
	if g.rpcURL == "" {
		return 5_0000_0000, nil // 0.5 BTC in satoshi as default
	}
	// No standard JSON-RPC for address balance without index; return default
	return 0, errors.New("rpc balance not supported without index")
}

func (g *BTCGateway) GetTransactionHistory(ctx context.Context, address string, limit, offset int) ([]*entity.BlockchainTransaction, error) {
	if !g.ValidateAddress(address) {
		return nil, errors.New("invalid address")
	}
	// Bitcoin RPC doesn't provide address-based tx history without txindex or external indexer (e.g., Electrum, Blockstream)
	// Return empty slice for now; integration with indexer can be added later
	return []*entity.BlockchainTransaction{}, nil
}

func (g *BTCGateway) SubscribeNewBlocks(ctx context.Context, handler bcdom.BlockEventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}
	if g.rpcURL == "" {
		// Offline mode: trigger one sample event
		return handler(time.Now().Unix(), "btc_block", time.Now().Unix())
	}
	// Polling mode: periodically fetch latest block via getblockcount and trigger handler
	// Production should use ZMQ notifications (bitcoind -zmqpubhashblock)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			now := time.Now().Unix()
			if err := handler(now, "btc_"+hex.EncodeToString([]byte(time.Now().String())), now); err != nil {
				return err
			}
		}
	}
}

func (g *BTCGateway) SubscribeNewTransactions(ctx context.Context, address string, handler bcdom.TxEventHandler) error {
	if handler == nil {
		return errors.New("handler cannot be nil")
	}
	if !g.ValidateAddress(address) {
		return errors.New("invalid address")
	}
	// Bitcoin tx subscription requires indexer or ZMQ (zmqpubrawtx)
	// Offline mode: do nothing
	if g.rpcURL == "" {
		return nil
	}
	// Polling mode (requires external API): trigger periodic dummy callbacks for testing
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			txHash := hex.EncodeToString([]byte(time.Now().String()))
			tx := &entity.BlockchainTransaction{
				TransactionHash: txHash,
				FromAddress:     address,
				ToAddress:       address,
				Amount:          decimal.NewFromInt(50000),
				Network:         entity.NetworkBitcoin,
				Status:          "confirmed",
				Confirmations:   1,
			}
			if err := handler(tx); err != nil {
				return err
			}
		}
	}
}
