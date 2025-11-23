package entity

import (
	"time"

	"github.com/google/uuid"
)

// User representa a entidade de usuário no domínio
type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
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
	ID               uuid.UUID
	UserID           uuid.UUID
	Address          string
	EncryptedPrivKey string
	Balance          float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
