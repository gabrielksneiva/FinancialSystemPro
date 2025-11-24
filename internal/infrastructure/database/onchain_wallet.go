package repositories

import (
	"time"

	"github.com/google/uuid"
)

// OnChainWallet representa uma carteira associada a um usuário para uma blockchain específica
type OnChainWallet struct {
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_blockchain"`
	Blockchain       string    `gorm:"type:text;not null;uniqueIndex:idx_user_blockchain"`
	Address          string    `gorm:"type:text;not null;uniqueIndex"`
	PublicKey        string    `gorm:"type:text;not null"`
	EncryptedPrivKey string    `gorm:"type:text;not null"`
}
