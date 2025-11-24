package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type User struct {
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Email     string    `gorm:"type:text;unique;not null"`
	Password  string    `gorm:"type:text;not null"`
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
}

type Account struct {
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Name      string    `gorm:"type:text;not null"`
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
}

type Balance struct {
	User      User            `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	Amount    decimal.Decimal `gorm:"type:numeric(15,2);not null"`
	ID        uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID       `gorm:"type:uuid;not null"`
}

type Transaction struct {
	User            User            `gorm:"foreignKey:AccountID;references:ID"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	TronTxHash      *string         `gorm:"type:text;nullable"`
	TronTxStatus    *string         `gorm:"type:text;default:null"`
	OnChainTxHash   *string         `gorm:"column:onchain_tx_hash;type:text;default:null"`
	OnChainTxStatus *string         `gorm:"column:onchain_tx_status;type:text;default:null"`
	OnChainChain    *string         `gorm:"column:onchain_chain;type:text;default:null"`
	Type            string          `gorm:"type:text;check:type IN ('deposit','withdraw','transfer')"`
	Category        string          `gorm:"type:text"`
	Description     string          `gorm:"type:text"`
	Amount          decimal.Decimal `gorm:"type:numeric(10,2)"`
	ID              uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AccountID       uuid.UUID       `gorm:"type:uuid;not null"`
}

type AuditLog struct {
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	Action     string         `gorm:"type:text"`
	IP         string         `gorm:"type:text"`
	OldPayload datatypes.JSON `gorm:"type:jsonb"`
	NewPayload datatypes.JSON `gorm:"type:jsonb"`
	ID         int            `gorm:"primaryKey;autoIncrement"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"`
}

// WalletInfo armazena informações de carteira TRON criptografada
type WalletInfo struct {
	User             User      `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
	TronAddress      string    `gorm:"type:text;not null;unique"`
	EncryptedPrivKey string    `gorm:"type:text;not null"`
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;unique"`
}
