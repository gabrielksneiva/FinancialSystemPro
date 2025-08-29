package workers

import (
	"financial-system-pro/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type JobType string

const (
	JobDeposit  JobType = "deposit"
	JobWithdraw JobType = "withdraw"
	JobTransfer JobType = "transfer"
)

type TransactionJob struct {
	Type        JobType
	Account     uuid.UUID
	Amount      decimal.Decimal
	ToEmail     string
	Req         any
	CallbackURL string
	JobID       uuid.UUID
}

type TransactionWorkerPool struct {
	DB      *repositories.NewDatabase
	Jobs    chan TransactionJob
	Workers int
	quit    chan struct{}
}

type JobResult struct {
	JobID   string `json:"job_id"`
	JobType string `json:"job_type"`
	Account string `json:"account"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}
