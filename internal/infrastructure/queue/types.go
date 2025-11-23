package workers

import (
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type JobType string

const (
	JobDeposit       JobType = "deposit"
	JobWithdraw      JobType = "withdraw"
	JobTransfer      JobType = "transfer"
	JobTronTxConfirm JobType = "tron_tx_confirm" // Monitorar confirmação de TX TRON
)

type TransactionJob struct {
	Type        JobType
	Amount      decimal.Decimal
	ToEmail     string
	CallbackURL string
	Account     uuid.UUID
	JobID       uuid.UUID
}

// TronTxConfirmJob monitora o status de uma transação TRON
type TronTxConfirmJob struct {
	Type          JobType
	TronTxHash    string
	CallbackURL   string
	CheckInterval int
	MaxChecks     int
	UserID        uuid.UUID
	TransactionID uuid.UUID
	JobID         uuid.UUID
}

type TransactionWorkerPool struct {
	DB      *repositories.NewDatabase
	Jobs    chan TransactionJob
	quit    chan struct{}
	Workers int
}

type JobResult struct {
	JobID   string `json:"job_id"`
	JobType string `json:"job_type"`
	Account string `json:"account"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}
