package events

import (
	"financial-system-pro/internal/shared/events"
	"time"

	"github.com/google/uuid"
)

// UserCreated evento disparado quando um usuário é criado
type UserCreated struct {
	events.BaseDomainEvent
	Email         string
	WalletAddress string
	WalletBalance float64
}

func NewUserCreated(userID uuid.UUID, email, walletAddress string) *UserCreated {
	return &UserCreated{
		BaseDomainEvent:     events.NewBaseDomainEvent("UserCreated", userID),
		Email:         email,
		WalletAddress: walletAddress,
		WalletBalance: 0.0,
	}
}

// UserDeactivated evento disparado quando um usuário é desativado
type UserDeactivated struct {
	events.BaseDomainEvent
	Reason string
}

func NewUserDeactivated(userID uuid.UUID, reason string) *UserDeactivated {
	return &UserDeactivated{
		BaseDomainEvent: events.NewBaseDomainEvent("UserDeactivated", userID),
		Reason:    reason,
	}
}

// UserActivated evento disparado quando um usuário é ativado
type UserActivated struct {
	events.BaseDomainEvent
}

func NewUserActivated(userID uuid.UUID) *UserActivated {
	return &UserActivated{
		BaseDomainEvent: events.NewBaseDomainEvent("UserActivated", userID),
	}
}

// PasswordChanged evento disparado quando senha é alterada
type PasswordChanged struct {
	events.BaseDomainEvent
	ChangedAt time.Time
}

func NewPasswordChanged(userID uuid.UUID) *PasswordChanged {
	return &PasswordChanged{
		BaseDomainEvent: events.NewBaseDomainEvent("PasswordChanged", userID),
		ChangedAt: time.Now(),
	}
}

// WalletCredited evento disparado quando fundos são creditados
type WalletCredited struct {
	events.BaseDomainEvent
	Amount     float64
	NewBalance float64
	Reference  string
}

func NewWalletCredited(userID uuid.UUID, amount, newBalance float64, reference string) *WalletCredited {
	return &WalletCredited{
		BaseDomainEvent:  events.NewBaseDomainEvent("WalletCredited", userID),
		Amount:     amount,
		NewBalance: newBalance,
		Reference:  reference,
	}
}

// WalletDebited evento disparado quando fundos são debitados
type WalletDebited struct {
	events.BaseDomainEvent
	Amount     float64
	NewBalance float64
	Reference  string
}

func NewWalletDebited(userID uuid.UUID, amount, newBalance float64, reference string) *WalletDebited {
	return &WalletDebited{
		BaseDomainEvent:  events.NewBaseDomainEvent("WalletDebited", userID),
		Amount:     amount,
		NewBalance: newBalance,
		Reference:  reference,
	}
}
