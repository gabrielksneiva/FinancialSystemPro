package entity

import (
	"time"

	"github.com/google/uuid"
)

// User representa a entidade de usuário no domínio
type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Password  string
	ID        uuid.UUID
}

// NewUser cria uma nova instância de User
func NewUser(email, password string) *User {
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Wallet representa a carteira associada ao usuário
type Wallet struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Address          string
	EncryptedPrivKey string
	Balance          float64
	ID               uuid.UUID
	UserID           uuid.UUID
}
