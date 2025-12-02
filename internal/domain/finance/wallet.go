package finance

import "errors"

// Wallet entity associated to an Account holding a single currency balance.
// Invariants: id != "", accountID != "".

type Wallet struct {
	id        string
	accountID string
	balance   Amount
}

func NewWallet(id, accountID string, initial Amount) (*Wallet, error) {
	if id == "" {
		return nil, errors.New("wallet id is required")
	}
	if accountID == "" {
		return nil, errors.New("account id is required")
	}
	return &Wallet{id: id, accountID: accountID, balance: initial}, nil
}

func (w *Wallet) ID() string        { return w.id }
func (w *Wallet) AccountID() string { return w.accountID }
func (w *Wallet) Balance() Amount   { return w.balance }

func (w *Wallet) Credit(a Amount) error {
	if w.balance.Currency() != a.Currency() {
		return errors.New("currency mismatch")
	}
	newBal, err := w.balance.Add(a)
	if err != nil {
		return err
	}
	w.balance = newBal
	return nil
}

func (w *Wallet) Debit(a Amount) error {
	if w.balance.Currency() != a.Currency() {
		return errors.New("currency mismatch")
	}
	newBal, err := w.balance.Sub(a)
	if err != nil {
		return err
	}
	w.balance = newBal
	return nil
}
