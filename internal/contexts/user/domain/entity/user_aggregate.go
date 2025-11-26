package entity

import (
	"errors"
	"financial-system-pro/internal/contexts/user/domain/valueobject"
	"time"

	"github.com/google/uuid"
)

// UserAggregate representa o agregado raiz User com sua Wallet
// Garante consistência transacional entre User e Wallet
type UserAggregate struct {
	user   *User
	wallet *Wallet
}

// User representa a entidade de usuário no domínio
type User struct {
	ID        uuid.UUID
	Email     valueobject.Email
	Password  valueobject.HashedPassword
	CreatedAt time.Time
	UpdatedAt time.Time
	isActive  bool
}

// NewUser cria uma nova instância de User
func NewUser(email valueobject.Email, password valueobject.HashedPassword) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
		isActive:  true,
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

// NewWallet cria uma nova wallet para um usuário
func NewWallet(userID uuid.UUID, address, encryptedPrivKey string) *Wallet {
	now := time.Now()
	return &Wallet{
		ID:               uuid.New(),
		UserID:           userID,
		Address:          address,
		EncryptedPrivKey: encryptedPrivKey,
		Balance:          0.0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewUserAggregate cria um novo agregado User com Wallet
func NewUserAggregate(email valueobject.Email, password valueobject.HashedPassword, address, encryptedPrivKey string) (*UserAggregate, error) {
	user := NewUser(email, password)
	wallet := NewWallet(user.ID, address, encryptedPrivKey)

	return &UserAggregate{
		user:   user,
		wallet: wallet,
	}, nil
}

// User retorna a entidade User
func (a *UserAggregate) User() *User {
	return a.user
}

// Wallet retorna a entidade Wallet
func (a *UserAggregate) Wallet() *Wallet {
	return a.wallet
}

// === Comportamentos do User ===

// ChangePassword altera a senha do usuário
func (u *User) ChangePassword(newPassword valueobject.HashedPassword) {
	u.Password = newPassword
	u.UpdatedAt = time.Now()
}

// Authenticate verifica se a senha fornecida é válida
func (u *User) Authenticate(rawPassword string) bool {
	return u.Password.Matches(rawPassword)
}

// Deactivate desativa o usuário
func (u *User) Deactivate() {
	u.isActive = false
	u.UpdatedAt = time.Now()
}

// Activate ativa o usuário
func (u *User) Activate() {
	u.isActive = true
	u.UpdatedAt = time.Now()
}

// IsActive verifica se o usuário está ativo
func (u *User) IsActive() bool {
	return u.isActive
}

// === Comportamentos da Wallet ===

// Credit adiciona fundos à wallet
func (w *Wallet) Credit(amount float64) error {
	if amount <= 0 {
		return errors.New("credit amount must be positive")
	}

	w.Balance += amount
	w.UpdatedAt = time.Now()
	return nil
}

// Debit remove fundos da wallet
func (w *Wallet) Debit(amount float64) error {
	if amount <= 0 {
		return errors.New("debit amount must be positive")
	}

	if w.Balance < amount {
		return errors.New("insufficient balance")
	}

	w.Balance -= amount
	w.UpdatedAt = time.Now()
	return nil
}

// HasSufficientBalance verifica se há saldo suficiente
func (w *Wallet) HasSufficientBalance(amount float64) bool {
	return w.Balance >= amount
}

// GetBalance retorna o saldo atual
func (w *Wallet) GetBalance() float64 {
	return w.Balance
}

// === Comportamentos do Agregado ===

// CreditWallet adiciona fundos à wallet do usuário
func (a *UserAggregate) CreditWallet(amount float64) error {
	if !a.user.IsActive() {
		return errors.New("cannot credit inactive user wallet")
	}
	return a.wallet.Credit(amount)
}

// DebitWallet remove fundos da wallet do usuário
func (a *UserAggregate) DebitWallet(amount float64) error {
	if !a.user.IsActive() {
		return errors.New("cannot debit inactive user wallet")
	}
	return a.wallet.Debit(amount)
}

// CanWithdraw verifica se o usuário pode fazer saque
func (a *UserAggregate) CanWithdraw(amount float64) bool {
	return a.user.IsActive() && a.wallet.HasSufficientBalance(amount)
}
