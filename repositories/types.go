package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string    `gorm:"type:text;unique;not null"`
	Password  string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Account struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Balance struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	User      User      `gorm:"foreignKey:UserID;references:ID"`
	Amount    float64   `gorm:"type:numeric(15,2);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Transaction struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AccountID        uuid.UUID  `gorm:"type:uuid;not null"`
	Type             string     `gorm:"type:text;check:type IN ('deposit','withdraw','transfer')"`
	Category         string     `gorm:"type:text"`
	Description      string     `gorm:"type:text"`
	Amount           float64    `gorm:"type:numeric(10,2)"`
	RelatedAccountID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
}

type AuditLog struct {
	ID         int            `gorm:"primaryKey;autoIncrement"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"`
	Action     string         `gorm:"type:text"`
	OldPayload datatypes.JSON `gorm:"type:jsonb"`
	NewPayload datatypes.JSON `gorm:"type:jsonb"`
	IP         string         `gorm:"type:text"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
}
