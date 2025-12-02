package finance

import "errors"

// Account aggregate root representing a financial owner entity.
// Invariants: id != "", holderID != "". Status transitions: Active -> Inactive -> Active.

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
)

type Account struct {
	id       string
	holderID string
	status   AccountStatus
}

func NewAccount(id, holderID string) (*Account, error) {
	if id == "" {
		return nil, errors.New("account id is required")
	}
	if holderID == "" {
		return nil, errors.New("holder id is required")
	}
	return &Account{id: id, holderID: holderID, status: AccountStatusActive}, nil
}

func (a *Account) ID() string            { return a.id }
func (a *Account) HolderID() string      { return a.holderID }
func (a *Account) Status() AccountStatus { return a.status }

func (a *Account) Deactivate() error {
	if a.status == AccountStatusInactive {
		return errors.New("account already inactive")
	}
	a.status = AccountStatusInactive
	return nil
}

func (a *Account) Activate() error {
	if a.status == AccountStatusActive {
		return errors.New("account already active")
	}
	a.status = AccountStatusActive
	return nil
}
