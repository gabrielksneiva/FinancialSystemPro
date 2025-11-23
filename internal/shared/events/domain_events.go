package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Eventos de Domínio - Transaction Context

// DepositCompletedEvent é publicado quando um depósito é concluído com sucesso
type DepositCompletedEvent struct {
	BaseEvent
	UserID   uuid.UUID       `json:"user_id"`
	Amount   decimal.Decimal `json:"amount"`
	TxHash   string          `json:"tx_hash"`
	Metadata map[string]any  `json:"metadata"`
}

func NewDepositCompletedEvent(userID uuid.UUID, amount decimal.Decimal, txHash string) DepositCompletedEvent {
	return DepositCompletedEvent{
		BaseEvent: NewBaseEvent("deposit.completed", userID.String()),
		UserID:    userID,
		Amount:    amount,
		TxHash:    txHash,
		Metadata:  make(map[string]any),
	}
}

// WithdrawCompletedEvent é publicado quando um saque é concluído
type WithdrawCompletedEvent struct {
	BaseEvent
	UserID   uuid.UUID       `json:"user_id"`
	Amount   decimal.Decimal `json:"amount"`
	TxHash   string          `json:"tx_hash"`
	Metadata map[string]any  `json:"metadata"`
}

func NewWithdrawCompletedEvent(userID uuid.UUID, amount decimal.Decimal, txHash string) WithdrawCompletedEvent {
	return WithdrawCompletedEvent{
		BaseEvent: NewBaseEvent("withdraw.completed", userID.String()),
		UserID:    userID,
		Amount:    amount,
		TxHash:    txHash,
		Metadata:  make(map[string]any),
	}
}

// TransferCompletedEvent é publicado quando uma transferência é concluída
type TransferCompletedEvent struct {
	BaseEvent
	FromUserID uuid.UUID       `json:"from_user_id"`
	ToUserID   uuid.UUID       `json:"to_user_id"`
	Amount     decimal.Decimal `json:"amount"`
	TxHash     string          `json:"tx_hash"`
	Metadata   map[string]any  `json:"metadata"`
}

func NewTransferCompletedEvent(fromUserID, toUserID uuid.UUID, amount decimal.Decimal, txHash string) TransferCompletedEvent {
	return TransferCompletedEvent{
		BaseEvent:  NewBaseEvent("transfer.completed", fromUserID.String()),
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		TxHash:     txHash,
		Metadata:   make(map[string]any),
	}
}

// TransactionFailedEvent é publicado quando uma transação falha
type TransactionFailedEvent struct {
	BaseEvent
	UserID    uuid.UUID       `json:"user_id"`
	TxType    string          `json:"tx_type"` // "deposit", "withdraw", "transfer"
	Amount    decimal.Decimal `json:"amount"`
	Reason    string          `json:"reason"`
	ErrorCode string          `json:"error_code"`
	Metadata  map[string]any  `json:"metadata"`
}

func NewTransactionFailedEvent(userID uuid.UUID, txType string, amount decimal.Decimal, reason, errorCode string) TransactionFailedEvent {
	return TransactionFailedEvent{
		BaseEvent: NewBaseEvent("transaction.failed", userID.String()),
		UserID:    userID,
		TxType:    txType,
		Amount:    amount,
		Reason:    reason,
		ErrorCode: errorCode,
		Metadata:  make(map[string]any),
	}
}

// Eventos de Domínio - User Context

// UserCreatedEvent é publicado quando um novo usuário é criado
type UserCreatedEvent struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserCreatedEvent(userID uuid.UUID, email, name string) UserCreatedEvent {
	return UserCreatedEvent{
		BaseEvent: NewBaseEvent("user.created", userID.String()),
		UserID:    userID,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// UserAuthenticatedEvent é publicado quando um usuário faz login
type UserAuthenticatedEvent struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

func NewUserAuthenticatedEvent(userID uuid.UUID, email, ipAddress, userAgent string) UserAuthenticatedEvent {
	return UserAuthenticatedEvent{
		BaseEvent: NewBaseEvent("user.authenticated", userID.String()),
		UserID:    userID,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

// Eventos de Domínio - Blockchain Context

// WalletCreatedEvent é publicado quando uma nova wallet é criada
type WalletCreatedEvent struct {
	BaseEvent
	UserID         uuid.UUID `json:"user_id"`
	WalletAddress  string    `json:"wallet_address"`
	BlockchainType string    `json:"blockchain_type"` // "TRON", "ETH", etc
}

func NewWalletCreatedEvent(userID uuid.UUID, walletAddress, blockchainType string) WalletCreatedEvent {
	return WalletCreatedEvent{
		BaseEvent:      NewBaseEvent("wallet.created", userID.String()),
		UserID:         userID,
		WalletAddress:  walletAddress,
		BlockchainType: blockchainType,
	}
}

// BlockchainTransactionConfirmedEvent é publicado quando uma tx blockchain é confirmada
type BlockchainTransactionConfirmedEvent struct {
	BaseEvent
	TxHash         string    `json:"tx_hash"`
	Confirmations  int       `json:"confirmations"`
	BlockNumber    int64     `json:"block_number"`
	ConfirmedAt    time.Time `json:"confirmed_at"`
	BlockchainType string    `json:"blockchain_type"`
}

func NewBlockchainTransactionConfirmedEvent(txHash string, confirmations int, blockNumber int64, blockchainType string) BlockchainTransactionConfirmedEvent {
	return BlockchainTransactionConfirmedEvent{
		BaseEvent:      NewBaseEvent("blockchain.transaction.confirmed", txHash),
		TxHash:         txHash,
		Confirmations:  confirmations,
		BlockNumber:    blockNumber,
		ConfirmedAt:    time.Now(),
		BlockchainType: blockchainType,
	}
}
