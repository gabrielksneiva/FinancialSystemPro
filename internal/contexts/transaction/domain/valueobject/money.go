package valueobject

import (
	"errors"

	"github.com/shopspring/decimal"
)

// Currency representa código ISO de moeda (ex: BRL, USD)
type Currency string

// Money VO com invariantes de não-negatividade.
type Money struct {
	amount   decimal.Decimal
	currency Currency
}

// NewMoney valida e constrói Money.
func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
	if currency == "" {
		return Money{}, errors.New("currency required")
	}
	if amount.IsNegative() {
		return Money{}, errors.New("amount cannot be negative")
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() decimal.Decimal { return m.amount }
func (m Money) Currency() Currency      { return m.currency }

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("currency mismatch")
	}
	return Money{amount: m.amount.Add(other.amount), currency: m.currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("currency mismatch")
	}
	res := m.amount.Sub(other.amount)
	if res.IsNegative() {
		return Money{}, errors.New("resulting amount negative")
	}
	return Money{amount: res, currency: m.currency}, nil
}
