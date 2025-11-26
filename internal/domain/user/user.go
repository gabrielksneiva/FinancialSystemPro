package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User aggregate root containing wallets and credentials.
type User struct {
	id        uuid.UUID
	email     Email
	password  HashedPassword
	createdAt time.Time
	updatedAt time.Time
	wallets   []uuid.UUID // references to wallet aggregate IDs (simplified)
}

// NewUser constructs a new User aggregate enforcing invariants.
func NewUser(email Email, password HashedPassword) *User {
	return &User{
		id:        uuid.New(),
		email:     email,
		password:  password,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		wallets:   []uuid.UUID{},
	}
}

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) Email() Email         { return u.email }
func (u *User) CreatedAt() time.Time { return u.createdAt }

// ChangePassword validates old password before updating.
func (u *User) ChangePassword(oldRaw, newRaw string) error {
	if !u.password.Matches(oldRaw) {
		return errors.New("invalid current password")
	}
	newHashed, err := HashPassword(newRaw)
	if err != nil {
		return err
	}
	u.password = newHashed
	u.updatedAt = time.Now()
	return nil
}

// AttachWallet registers a wallet ID with this user (no duplicates).
func (u *User) AttachWallet(walletID uuid.UUID) error {
	for _, w := range u.wallets {
		if w == walletID {
			return errors.New("wallet already attached")
		}
	}
	u.wallets = append(u.wallets, walletID)
	u.updatedAt = time.Now()
	return nil
}
