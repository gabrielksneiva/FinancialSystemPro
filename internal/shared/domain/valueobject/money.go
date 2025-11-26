package valueobject

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// Currency representa uma moeda
type Currency string

const (
	BRL Currency = "BRL"
	USD Currency = "USD"
	EUR Currency = "EUR"
)

// Money é um Value Object que representa dinheiro com validação
type Money struct {
	amount   decimal.Decimal
	currency Currency
}

// NewMoney cria uma nova instância de Money com validação
func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
	if amount.IsNegative() {
		return Money{}, errors.New("amount cannot be negative")
	}

	if currency == "" {
		return Money{}, errors.New("currency is required")
	}

	// Validar moeda suportada
	switch currency {
	case BRL, USD, EUR:
		// OK
	default:
		return Money{}, fmt.Errorf("unsupported currency: %s", currency)
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

// MustNewMoney cria Money e entra em panic se inválido (use apenas em testes/setup)
func MustNewMoney(amount decimal.Decimal, currency Currency) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero retorna um Money com valor zero na moeda especificada
func Zero(currency Currency) Money {
	return Money{
		amount:   decimal.Zero,
		currency: currency,
	}
}

// Amount retorna o valor monetário
func (m Money) Amount() decimal.Decimal {
	return m.amount
}

// Currency retorna a moeda
func (m Money) Currency() Currency {
	return m.currency
}

// IsZero verifica se o valor é zero
func (m Money) IsZero() bool {
	return m.amount.IsZero()
}

// IsPositive verifica se o valor é positivo
func (m Money) IsPositive() bool {
	return m.amount.IsPositive()
}

// Add soma dois valores monetários (devem ter a mesma moeda)
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.currency, other.currency)
	}

	return Money{
		amount:   m.amount.Add(other.amount),
		currency: m.currency,
	}, nil
}

// Subtract subtrai dois valores monetários (devem ter a mesma moeda)
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.currency, other.currency)
	}

	result := m.amount.Sub(other.amount)
	if result.IsNegative() {
		return Money{}, errors.New("result cannot be negative")
	}

	return Money{
		amount:   result,
		currency: m.currency,
	}, nil
}

// Multiply multiplica o valor por um fator
func (m Money) Multiply(factor decimal.Decimal) (Money, error) {
	result := m.amount.Mul(factor)
	if result.IsNegative() {
		return Money{}, errors.New("result cannot be negative")
	}

	return Money{
		amount:   result,
		currency: m.currency,
	}, nil
}

// GreaterThan verifica se é maior que outro Money
func (m Money) GreaterThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", m.currency, other.currency)
	}
	return m.amount.GreaterThan(other.amount), nil
}

// GreaterThanOrEqual verifica se é maior ou igual a outro Money
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", m.currency, other.currency)
	}
	return m.amount.GreaterThanOrEqual(other.amount), nil
}

// LessThan verifica se é menor que outro Money
func (m Money) LessThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, fmt.Errorf("cannot compare different currencies: %s and %s", m.currency, other.currency)
	}
	return m.amount.LessThan(other.amount), nil
}

// Equals verifica igualdade
func (m Money) Equals(other Money) bool {
	return m.currency == other.currency && m.amount.Equal(other.amount)
}

// String retorna representação em string
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.currency, m.amount.StringFixed(2))
}
