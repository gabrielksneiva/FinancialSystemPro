package cqrs

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// UserReadModel represents the read-optimized user view for queries
type UserReadModel struct {
	ID                     uuid.UUID       `json:"id" db:"id"`
	Email                  string          `json:"email" db:"email"`
	IsActive               bool            `json:"is_active" db:"is_active"`
	WalletAddress          string          `json:"wallet_address" db:"wallet_address"`
	Balance                decimal.Decimal `json:"balance" db:"balance"`
	TotalDeposits          decimal.Decimal `json:"total_deposits" db:"total_deposits"`
	TotalWithdrawals       decimal.Decimal `json:"total_withdrawals" db:"total_withdrawals"`
	TotalTransfersSent     decimal.Decimal `json:"total_transfers_sent" db:"total_transfers_sent"`
	TotalTransfersReceived decimal.Decimal `json:"total_transfers_received" db:"total_transfers_received"`
	TransactionCount       int             `json:"transaction_count" db:"transaction_count"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
	LastTransactionAt      *time.Time      `json:"last_transaction_at,omitempty" db:"last_transaction_at"`
}

// TransactionReadModel represents the read-optimized transaction view
type TransactionReadModel struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	Type          string          `json:"type" db:"type"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	Status        string          `json:"status" db:"status"`
	Blockchain    string          `json:"blockchain,omitempty" db:"blockchain"`
	TxHash        string          `json:"tx_hash,omitempty" db:"tx_hash"`
	FromAddress   string          `json:"from_address,omitempty" db:"from_address"`
	ToAddress     string          `json:"to_address,omitempty" db:"to_address"`
	Confirmations int             `json:"confirmations" db:"confirmations"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// UserQuery represents a query for users
type UserQuery struct {
	ID         *uuid.UUID       `json:"id,omitempty"`
	Email      *string          `json:"email,omitempty"`
	IsActive   *bool            `json:"is_active,omitempty"`
	MinBalance *decimal.Decimal `json:"min_balance,omitempty"`
	Limit      int              `json:"limit,omitempty"`
	Offset     int              `json:"offset,omitempty"`
}

// TransactionQuery represents a query for transactions
type TransactionQuery struct {
	ID        *uuid.UUID       `json:"id,omitempty"`
	UserID    *uuid.UUID       `json:"user_id,omitempty"`
	Type      *string          `json:"type,omitempty"`
	Status    *string          `json:"status,omitempty"`
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"`
	FromDate  *time.Time       `json:"from_date,omitempty"`
	ToDate    *time.Time       `json:"to_date,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// UserStatistics aggregated stats for a user
type UserStatistics struct {
	UserID                 uuid.UUID       `json:"user_id"`
	TotalDeposits          decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals       decimal.Decimal `json:"total_withdrawals"`
	TotalTransfersSent     decimal.Decimal `json:"total_transfers_sent"`
	TotalTransfersReceived decimal.Decimal `json:"total_transfers_received"`
	NetFlow                decimal.Decimal `json:"net_flow"`
	TransactionCount       int             `json:"transaction_count"`
	AverageTransactionSize decimal.Decimal `json:"average_transaction_size"`
	LastActivityAt         *time.Time      `json:"last_activity_at,omitempty"`
}
