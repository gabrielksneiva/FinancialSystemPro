package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionModel representa a tabela de transações no banco de dados (GORM)
// Separado da entidade de domínio para manter Clean Architecture
type TransactionModel struct {
	ID             uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	Amount         decimal.Decimal `gorm:"type:numeric(15,2);not null"`
	Type           string          `gorm:"type:text;not null"`
	Status         string          `gorm:"type:text;not null"`
	Hash           string          `gorm:"type:text;unique"`
	FromAddress    string          `gorm:"type:text"`
	ToAddress      string          `gorm:"type:text"`
	BlockchainType string          `gorm:"type:text"`
	Confirmations  int             `gorm:"type:int;default:0"`
	Fee            decimal.Decimal `gorm:"type:numeric(15,2);default:0"`
	Metadata       string          `gorm:"type:jsonb"`
	CreatedAt      time.Time       `gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime"`
	CompletedAt    *time.Time      `gorm:"type:timestamp"`
}

// TableName sobrescreve o nome da tabela para manter compatibilidade
func (TransactionModel) TableName() string {
	return "transaction_context.transactions"
}
