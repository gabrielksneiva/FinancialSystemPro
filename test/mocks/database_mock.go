package mocks

import (
	"errors"
	r "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of the database interface
type MockDatabase struct {
	mock.Mock
	ShouldError    bool
	MockWalletInfo *r.WalletInfo
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

func (m *MockDatabase) FindUserByField(field string, value interface{}) (*r.User, error) {
	if m.ShouldError {
		return nil, errors.New("database error")
	}
	return &r.User{
		ID:    uuid.New(),
		Email: "test@test.com",
	}, nil
}

func (m *MockDatabase) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	if m.ShouldError {
		return decimal.Zero, errors.New("database error")
	}
	return decimal.NewFromInt(100), nil
}

func (m *MockDatabase) Transaction(userID uuid.UUID, amount decimal.Decimal, txType string) error {
	if m.ShouldError {
		return errors.New("database error")
	}
	return nil
}

func (m *MockDatabase) Insert(value interface{}) error {
	if m.ShouldError {
		return errors.New("database error")
	}
	return nil
}

func (m *MockDatabase) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	if m.ShouldError {
		return errors.New("database error")
	}
	return nil
}

func (m *MockDatabase) GetWalletInfo(userID uuid.UUID) (*r.WalletInfo, error) {
	if m.ShouldError || m.MockWalletInfo == nil {
		return nil, errors.New("wallet not found")
	}
	return m.MockWalletInfo, nil
}

func (m *MockDatabase) GetWalletInfoByAddress(address string) (*r.WalletInfo, error) {
	if m.ShouldError {
		return nil, errors.New("wallet not found")
	}
	return &r.WalletInfo{
		TronAddress: address,
	}, nil
}

func (m *MockDatabase) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	if m.ShouldError {
		return errors.New("database error")
	}
	return nil
}
