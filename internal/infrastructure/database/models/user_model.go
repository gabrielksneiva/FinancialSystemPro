package models

import (
	"time"

	"github.com/google/uuid"
)

// UserModel representa a tabela de usuários no banco de dados (GORM)
// Separado da entidade de domínio para manter Clean Architecture
type UserModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string    `gorm:"type:text;unique;not null"`
	Password  string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName sobrescreve o nome da tabela para manter compatibilidade
func (UserModel) TableName() string {
	return "user_context.users"
}

// WalletModel representa a tabela de carteiras no banco de dados (GORM)
type WalletModel struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	Address          string    `gorm:"type:text;unique;not null"`
	EncryptedPrivKey string    `gorm:"type:text;not null"`
	Balance          float64   `gorm:"type:numeric(15,2);default:0"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// TableName sobrescreve o nome da tabela para manter compatibilidade
func (WalletModel) TableName() string {
	return "user_context.wallets"
}
