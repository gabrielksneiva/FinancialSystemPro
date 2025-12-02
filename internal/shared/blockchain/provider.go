package blockchain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// ChainType identifica o tipo de blockchain
type ChainType string

const (
	ChainEthereum ChainType = "ethereum"
	ChainBitcoin  ChainType = "bitcoin"
	ChainSolana   ChainType = "solana"
	ChainTron     ChainType = "tron"
	ChainTON      ChainType = "ton"
	ChainSUI      ChainType = "sui"
	ChainAptos    ChainType = "aptos"
	ChainCosmos   ChainType = "cosmos"
)

// NetworkType identifica o tipo de rede
type NetworkType string

const (
	NetworkMainnet NetworkType = "mainnet"
	NetworkTestnet NetworkType = "testnet"
	NetworkDevnet  NetworkType = "devnet"
)

// Wallet representa uma carteira blockchain gerada
type Wallet struct {
	Address    string    `json:"address"`
	PublicKey  string    `json:"public_key"`
	PrivateKey string    `json:"private_key,omitempty"` // Opcional, pode ser vazio por segurança
	ChainType  ChainType `json:"chain_type"`
	CreatedAt  time.Time `json:"created_at"`
}

// Balance representa o saldo de uma carteira
type Balance struct {
	Address       string          `json:"address"`
	Amount        decimal.Decimal `json:"amount"`
	AmountRaw     string          `json:"amount_raw"` // Valor em unidade mínima (wei, satoshi, lamport)
	Currency      string          `json:"currency"`   // ETH, BTC, SOL, TRX
	Decimals      int             `json:"decimals"`
	BlockNumber   int64           `json:"block_number"`
	LastUpdatedAt time.Time       `json:"last_updated_at"`
}

// TransactionIntent representa a intenção de criar uma transação
type TransactionIntent struct {
	From      string          `json:"from"`
	To        string          `json:"to"`
	Amount    decimal.Decimal `json:"amount"`
	AmountRaw string          `json:"amount_raw,omitempty"` // Se fornecido, usa este ao invés de Amount
	Data      []byte          `json:"data,omitempty"`       // Para smart contracts
	Nonce     *uint64         `json:"nonce,omitempty"`
	GasLimit  *uint64         `json:"gas_limit,omitempty"`
	Memo      string          `json:"memo,omitempty"` // Para blockchains que suportam memo
}

// FeeEstimate representa a estimativa de taxa de transação
type FeeEstimate struct {
	ChainType   ChainType       `json:"chain_type"`
	Low         decimal.Decimal `json:"low"`                 // Estimativa baixa (lenta)
	Medium      decimal.Decimal `json:"medium"`              // Estimativa média (normal)
	High        decimal.Decimal `json:"high"`                // Estimativa alta (rápida)
	Currency    string          `json:"currency"`            // Moeda da taxa (ETH, BTC, etc)
	GasPrice    *string         `json:"gas_price,omitempty"` // Para chains que usam gas
	GasLimit    *uint64         `json:"gas_limit,omitempty"` // Para chains que usam gas
	EstimatedAt time.Time       `json:"estimated_at"`
}

// UnsignedTransaction representa uma transação não assinada
type UnsignedTransaction struct {
	ChainType ChainType              `json:"chain_type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Amount    decimal.Decimal        `json:"amount"`
	Fee       decimal.Decimal        `json:"fee"`
	Nonce     uint64                 `json:"nonce"`
	Data      []byte                 `json:"data,omitempty"`
	RawData   map[string]interface{} `json:"raw_data"` // Dados específicos da chain
	CreatedAt time.Time              `json:"created_at"`
}

// SignedTransaction representa uma transação assinada pronta para broadcast
type SignedTransaction struct {
	ChainType ChainType              `json:"chain_type"`
	RawTx     string                 `json:"raw_tx"`  // Transação serializada
	TxHash    string                 `json:"tx_hash"` // Hash da transação
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Amount    decimal.Decimal        `json:"amount"`
	Fee       decimal.Decimal        `json:"fee"`
	Signature string                 `json:"signature"`
	RawData   map[string]interface{} `json:"raw_data"` // Dados específicos da chain
	SignedAt  time.Time              `json:"signed_at"`
}

// TransactionReceipt representa o recibo de uma transação transmitida
type TransactionReceipt struct {
	TxHash        string          `json:"tx_hash"`
	ChainType     ChainType       `json:"chain_type"`
	Status        TxStatus        `json:"status"`
	BlockNumber   *int64          `json:"block_number,omitempty"`
	BlockHash     *string         `json:"block_hash,omitempty"`
	From          string          `json:"from"`
	To            string          `json:"to"`
	Amount        decimal.Decimal `json:"amount"`
	Fee           decimal.Decimal `json:"fee"`
	GasUsed       *uint64         `json:"gas_used,omitempty"`
	Confirmations int             `json:"confirmations"`
	BroadcastAt   time.Time       `json:"broadcast_at"`
	ConfirmedAt   *time.Time      `json:"confirmed_at,omitempty"`
	RawReceipt    json.RawMessage `json:"raw_receipt,omitempty"`
}

// TxStatus representa o status de uma transação
type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
	TxStatusUnknown   TxStatus = "unknown"
)

// TransactionStatus representa o status detalhado de uma transação
type TransactionStatus struct {
	TxHash        string     `json:"tx_hash"`
	Status        TxStatus   `json:"status"`
	Confirmations int        `json:"confirmations"`
	BlockNumber   *int64     `json:"block_number,omitempty"`
	BlockHash     *string    `json:"block_hash,omitempty"`
	Timestamp     *time.Time `json:"timestamp,omitempty"`
	Error         *string    `json:"error,omitempty"`
}

// PaginationOptions representa opções de paginação
type PaginationOptions struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Cursor string `json:"cursor,omitempty"` // Para cursor-based pagination
}

// TransactionHistory representa o histórico de transações
type TransactionHistory struct {
	Address      string                  `json:"address"`
	Transactions []HistoricalTransaction `json:"transactions"`
	Total        int                     `json:"total"`
	HasMore      bool                    `json:"has_more"`
	NextCursor   string                  `json:"next_cursor,omitempty"`
}

// HistoricalTransaction representa uma transação histórica
type HistoricalTransaction struct {
	TxHash        string          `json:"tx_hash"`
	From          string          `json:"from"`
	To            string          `json:"to"`
	Amount        decimal.Decimal `json:"amount"`
	Fee           decimal.Decimal `json:"fee"`
	Status        TxStatus        `json:"status"`
	BlockNumber   int64           `json:"block_number"`
	Confirmations int             `json:"confirmations"`
	Timestamp     time.Time       `json:"timestamp"`
	Type          string          `json:"type"` // sent, received, contract_call
}

// Block representa um bloco blockchain
type Block struct {
	Number     int64     `json:"number"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parent_hash"`
	Timestamp  time.Time `json:"timestamp"`
	TxCount    int       `json:"tx_count"`
	Miner      string    `json:"miner,omitempty"`
	Difficulty *string   `json:"difficulty,omitempty"`
	TotalDiff  *string   `json:"total_difficulty,omitempty"`
	Size       int64     `json:"size"`
	GasUsed    *uint64   `json:"gas_used,omitempty"`
	GasLimit   *uint64   `json:"gas_limit,omitempty"`
}

// BlockHandler é um callback para novos blocos
type BlockHandler func(ctx context.Context, block *Block) error

// TxFilter representa filtros para transações
type TxFilter struct {
	Addresses []string `json:"addresses,omitempty"`
	FromBlock *int64   `json:"from_block,omitempty"`
	ToBlock   *int64   `json:"to_block,omitempty"`
}

// TxHandler é um callback para novas transações
type TxHandler func(ctx context.Context, tx *HistoricalTransaction) error

// PrivateKey representa uma chave privada (wrapper seguro)
type PrivateKey struct {
	Raw []byte // Nunca expor diretamente em JSON
}

// MarshalJSON implementa json.Marshaler para segurança
func (pk PrivateKey) MarshalJSON() ([]byte, error) {
	return []byte(`"***REDACTED***"`), nil
}

// Provider define a interface unificada para qualquer blockchain
type Provider interface {
	// Metadata retorna informações sobre o provider
	ChainType() ChainType
	NetworkType() NetworkType
	IsHealthy(ctx context.Context) bool

	// Wallet operations
	GenerateWallet(ctx context.Context) (*Wallet, error)
	ValidateAddress(address string) error
	ImportWallet(ctx context.Context, privateKey string) (*Wallet, error)

	// Balance operations
	GetBalance(ctx context.Context, address string) (*Balance, error)

	// Transaction operations
	EstimateFee(ctx context.Context, intent *TransactionIntent) (*FeeEstimate, error)
	BuildTransaction(ctx context.Context, intent *TransactionIntent) (*UnsignedTransaction, error)
	SignTransaction(ctx context.Context, tx *UnsignedTransaction, privateKey *PrivateKey) (*SignedTransaction, error)
	BroadcastTransaction(ctx context.Context, tx *SignedTransaction) (*TransactionReceipt, error)
	GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error)

	// History
	GetTransactionHistory(ctx context.Context, address string, opts *PaginationOptions) (*TransactionHistory, error)

	// Monitoring (opcional - nem todas as chains suportam)
	SubscribeNewBlocks(ctx context.Context, handler BlockHandler) error
	SubscribeNewTransactions(ctx context.Context, filter *TxFilter, handler TxHandler) error
	UnsubscribeAll(ctx context.Context) error
}

// Registry gerencia múltiplos providers
type Registry interface {
	Register(provider Provider) error
	Get(chainType ChainType) (Provider, error)
	List() []Provider
	Exists(chainType ChainType) bool
	Unregister(chainType ChainType) error
	Count() int
}

// ProviderCapabilities indica recursos suportados por um provider
type ProviderCapabilities struct {
	SupportsSmartContracts   bool
	SupportsTokens           bool
	SupportsStaking          bool
	SupportsSubscriptions    bool
	SupportsMemo             bool
	SupportsMultiSig         bool
	RequiresGas              bool
	NativeTokenDecimals      int
	MinConfirmationsRequired int
	AverageBlockTime         time.Duration
}

// CapableProvider estende Provider com capabilities
type CapableProvider interface {
	Provider
	GetCapabilities() *ProviderCapabilities
}
