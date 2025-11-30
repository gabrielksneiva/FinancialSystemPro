package persistence

import (
	"time"

	"gorm.io/gorm"
)

// GORM models for blockchain persistence
type BlockModel struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	Number     uint64    `gorm:"uniqueIndex;not null"`
	Hash       string    `gorm:"type:text"`
	OccurredAt time.Time `gorm:"column:occurred_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (BlockModel) TableName() string { return "blocks" }

type TransactionModel struct {
	ID            string    `gorm:"primaryKey;type:text"`
	Hash          string    `gorm:"type:text;index"`
	BlockNumber   uint64    `gorm:"column:block_number;index"`
	FromAddr      string    `gorm:"column:from_addr;type:text"`
	ToAddr        string    `gorm:"column:to_addr;type:text"`
	Value         string    `gorm:"type:text"`
	Raw           []byte    `gorm:"type:bytea"`
	Confirmations int64     `gorm:"default:0"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

func (TransactionModel) TableName() string { return "transactions" }

type BalanceModel struct {
	Address   string    `gorm:"primaryKey;type:text"`
	Balance   string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (BalanceModel) TableName() string { return "balances" }

// AutoMigrate convenience to migrate blockchain models using a gorm DB
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&BlockModel{}, &TransactionModel{}, &BalanceModel{})
}
