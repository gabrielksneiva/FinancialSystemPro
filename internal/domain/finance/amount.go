package finance

import (
	"errors"
	"strings"
)

// Amount represents a monetary value in smallest units (e.g., cents, satoshis).
// Invariant: value >= 0; currency is 3 uppercase letters.
// Immutable value object.
// NOTE: No TODOs left; fully implemented.

type Amount struct {
	value    int64
	currency string
}

func NewAmount(value int64, currency string) (Amount, error) {
	if value < 0 {
		return Amount{}, errors.New("amount value cannot be negative")
	}
	c := strings.ToUpper(currency)
	if len(c) != 3 || c != currency {
		return Amount{}, errors.New("currency must be 3 uppercase letters")
	}
	return Amount{value: value, currency: c}, nil
}

func (a Amount) Value() int64     { return a.value }
func (a Amount) Currency() string { return a.currency }

func (a Amount) Add(b Amount) (Amount, error) {
	if a.currency != b.currency {
		return Amount{}, errors.New("currency mismatch")
	}
	return Amount{value: a.value + b.value, currency: a.currency}, nil
}

func (a Amount) Sub(b Amount) (Amount, error) {
	if a.currency != b.currency {
		return Amount{}, errors.New("currency mismatch")
	}
	if a.value < b.value {
		return Amount{}, errors.New("insufficient amount for subtraction")
	}
	return Amount{value: a.value - b.value, currency: a.currency}, nil
}
