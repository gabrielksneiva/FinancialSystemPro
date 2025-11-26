package valueobject

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewMoney_ValidInput(t *testing.T) {
	amount := decimal.NewFromFloat(100.50)
	money, err := NewMoney(amount, "BRL")

	assert.NoError(t, err)
	assert.Equal(t, amount, money.Amount())
	assert.Equal(t, Currency("BRL"), money.Currency())
}

func TestNewMoney_NegativeAmount(t *testing.T) {
	amount := decimal.NewFromFloat(-10.0)
	_, err := NewMoney(amount, "USD")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
}

func TestNewMoney_EmptyCurrency(t *testing.T) {
	amount := decimal.NewFromInt(50)
	_, err := NewMoney(amount, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency required")
}

func TestNewMoney_ZeroAmount(t *testing.T) {
	amount := decimal.Zero
	money, err := NewMoney(amount, "EUR")

	assert.NoError(t, err)
	assert.True(t, money.Amount().IsZero())
	assert.Equal(t, Currency("EUR"), money.Currency())
}

func TestMoney_Add_SameCurrency(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(100), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(50), "BRL")

	result, err := m1.Add(m2)

	assert.NoError(t, err)
	assert.Equal(t, "150", result.Amount().String())
	assert.Equal(t, Currency("BRL"), result.Currency())
}

func TestMoney_Add_DifferentCurrency(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(100), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(50), "USD")

	_, err := m1.Add(m2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency mismatch")
}

func TestMoney_Sub_ValidOperation(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(100), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(30), "BRL")

	result, err := m1.Sub(m2)

	assert.NoError(t, err)
	assert.Equal(t, "70", result.Amount().String())
	assert.Equal(t, Currency("BRL"), result.Currency())
}

func TestMoney_Sub_NegativeResult(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(50), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(100), "BRL")

	_, err := m1.Sub(m2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resulting amount negative")
}

func TestMoney_Sub_DifferentCurrency(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(100), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(30), "EUR")

	_, err := m1.Sub(m2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency mismatch")
}

func TestMoney_Sub_ResultZero(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(100), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(100), "BRL")

	result, err := m1.Sub(m2)

	assert.NoError(t, err)
	assert.True(t, result.Amount().IsZero())
	assert.Equal(t, Currency("BRL"), result.Currency())
}

func TestCurrency_Type(t *testing.T) {
	c := Currency("USD")
	assert.Equal(t, "USD", string(c))
}

func TestMoney_Getters(t *testing.T) {
	amount := decimal.NewFromFloat(123.45)
	money, _ := NewMoney(amount, "JPY")

	assert.Equal(t, amount, money.Amount())
	assert.Equal(t, Currency("JPY"), money.Currency())
}

func TestMoney_AddMultiple(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromInt(10), "BRL")
	m2, _ := NewMoney(decimal.NewFromInt(20), "BRL")
	m3, _ := NewMoney(decimal.NewFromInt(30), "BRL")

	result, err := m1.Add(m2)
	assert.NoError(t, err)
	result, err = result.Add(m3)
	assert.NoError(t, err)

	assert.Equal(t, "60", result.Amount().String())
}

func TestMoney_Decimals(t *testing.T) {
	m1, _ := NewMoney(decimal.NewFromFloat(10.99), "BRL")
	m2, _ := NewMoney(decimal.NewFromFloat(5.01), "BRL")

	result, err := m1.Add(m2)

	assert.NoError(t, err)
	assert.Equal(t, "16", result.Amount().String())
}
