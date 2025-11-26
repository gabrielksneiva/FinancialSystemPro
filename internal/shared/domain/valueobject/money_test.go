package valueobject

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		amount      decimal.Decimal
		currency    Currency
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid BRL amount",
			amount:   decimal.NewFromFloat(100.50),
			currency: BRL,
			wantErr:  false,
		},
		{
			name:     "valid USD amount",
			amount:   decimal.NewFromFloat(1000.99),
			currency: USD,
			wantErr:  false,
		},
		{
			name:     "valid EUR amount",
			amount:   decimal.NewFromFloat(500.00),
			currency: EUR,
			wantErr:  false,
		},
		{
			name:        "negative amount",
			amount:      decimal.NewFromFloat(-50.00),
			currency:    BRL,
			wantErr:     true,
			errContains: "amount cannot be negative",
		},
		{
			name:        "invalid currency",
			amount:      decimal.NewFromFloat(100.00),
			currency:    Currency("JPY"),
			wantErr:     true,
			errContains: "unsupported currency",
		},
		{
			name:        "empty currency",
			amount:      decimal.NewFromFloat(100.00),
			currency:    Currency(""),
			wantErr:     true,
			errContains: "currency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoney(tt.amount, tt.currency)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewMoney() expected error containing '%s', got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewMoney() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("NewMoney() unexpected error = %v", err)
					return
				}
				if money.Currency() != tt.currency {
					t.Errorf("NewMoney() currency = %v, want %v", money.Currency(), tt.currency)
				}
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name        string
		money1      Money
		money2      Money
		wantAmount  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "add same currency",
			money1:     mustNewMoney(decimal.NewFromFloat(100.50), BRL),
			money2:     mustNewMoney(decimal.NewFromFloat(50.25), BRL),
			wantAmount: "150.75",
			wantErr:    false,
		},
		{
			name:        "add different currencies",
			money1:      mustNewMoney(decimal.NewFromFloat(100.00), BRL),
			money2:      mustNewMoney(decimal.NewFromFloat(50.00), USD),
			wantErr:     true,
			errContains: "cannot add different currencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.money1.Add(tt.money2)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Add() expected error containing '%s', got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Add() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Add() unexpected error = %v", err)
					return
				}
				if result.Amount().String() != tt.wantAmount {
					t.Errorf("Add() amount = %v, want %v", result.Amount().String(), tt.wantAmount)
				}
			}
		})
	}
}

func TestMoney_Subtract(t *testing.T) {
	tests := []struct {
		name        string
		money1      Money
		money2      Money
		wantAmount  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "subtract same currency",
			money1:     mustNewMoney(decimal.NewFromFloat(100.50), BRL),
			money2:     mustNewMoney(decimal.NewFromFloat(50.25), BRL),
			wantAmount: "50.25",
			wantErr:    false,
		},
		{
			name:        "subtract different currencies",
			money1:      mustNewMoney(decimal.NewFromFloat(100.00), BRL),
			money2:      mustNewMoney(decimal.NewFromFloat(50.00), USD),
			wantErr:     true,
			errContains: "cannot subtract",
		},
		{
			name:        "subtract resulting in negative",
			money1:      mustNewMoney(decimal.NewFromFloat(50.00), BRL),
			money2:      mustNewMoney(decimal.NewFromFloat(100.00), BRL),
			wantErr:     true,
			errContains: "result cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.money1.Subtract(tt.money2)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Subtract() expected error containing '%s', got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Subtract() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Subtract() unexpected error = %v", err)
					return
				}
				if result.Amount().String() != tt.wantAmount {
					t.Errorf("Subtract() amount = %v, want %v", result.Amount().String(), tt.wantAmount)
				}
			}
		})
	}
}

func TestMoney_Multiply(t *testing.T) {
	tests := []struct {
		name        string
		money       Money
		multiplier  decimal.Decimal
		wantAmount  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "multiply by positive number",
			money:      mustNewMoney(decimal.NewFromFloat(100.00), BRL),
			multiplier: decimal.NewFromInt(2),
			wantAmount: "200",
			wantErr:    false,
		},
		{
			name:       "multiply by decimal",
			money:      mustNewMoney(decimal.NewFromFloat(100.00), BRL),
			multiplier: decimal.NewFromFloat(1.5),
			wantAmount: "150",
			wantErr:    false,
		},
		{
			name:        "multiply by negative number",
			money:       mustNewMoney(decimal.NewFromFloat(100.00), BRL),
			multiplier:  decimal.NewFromInt(-2),
			wantErr:     true,
			errContains: "result cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.money.Multiply(tt.multiplier)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Multiply() expected error containing '%s', got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Multiply() error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Multiply() unexpected error = %v", err)
					return
				}
				if result.Amount().String() != tt.wantAmount {
					t.Errorf("Multiply() amount = %v, want %v", result.Amount().String(), tt.wantAmount)
				}
			}
		})
	}
}

func TestMoney_Equals(t *testing.T) {
	tests := []struct {
		name   string
		money1 Money
		money2 Money
		want   bool
	}{
		{
			name:   "equal money",
			money1: mustNewMoney(decimal.NewFromFloat(100.50), BRL),
			money2: mustNewMoney(decimal.NewFromFloat(100.50), BRL),
			want:   true,
		},
		{
			name:   "different amounts",
			money1: mustNewMoney(decimal.NewFromFloat(100.50), BRL),
			money2: mustNewMoney(decimal.NewFromFloat(100.51), BRL),
			want:   false,
		},
		{
			name:   "different currencies",
			money1: mustNewMoney(decimal.NewFromFloat(100.50), BRL),
			money2: mustNewMoney(decimal.NewFromFloat(100.50), USD),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.money1.Equals(tt.money2); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions
func mustNewMoney(amount decimal.Decimal, currency Currency) Money {
	money, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return money
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
