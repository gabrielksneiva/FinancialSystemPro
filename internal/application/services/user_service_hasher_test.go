package services

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockDBUser minimal for user creation flow
type mockDBUser struct{ inserted *repositories.User }

func (m *mockDBUser) FindUserByField(field string, value any) (*repositories.User, error) {
	return nil, nil
}
func (m *mockDBUser) Insert(v any) error {
	if u, ok := v.(*repositories.User); ok {
		m.inserted = u
	}
	return nil
}
func (m *mockDBUser) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return nil
}
func (m *mockDBUser) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return nil, errors.NewDatabaseError("not implemented", nil)
}
func (m *mockDBUser) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return nil
}
func (m *mockDBUser) Balance(userID uuid.UUID) (decimal.Decimal, error) { return decimal.Zero, nil }
func (m *mockDBUser) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type stubHasher struct{}

func (stubHasher) Compare(raw, hashed string) (bool, error) { return raw == hashed, nil }
func (stubHasher) Hash(raw string) (string, error)          { return "PREFIX_" + raw, nil }

func TestUserService_CreateNewUser_CustomHasher(t *testing.T) {
	db := &mockDBUser{}
	svc := NewUserService(db, zap.NewNop(), entities.WalletManager(nil)).WithPasswordHasher(stubHasher{})
	appErr := svc.CreateNewUser(&dto.UserRequest{Email: "x@y.com", Password: "pw"})
	require.Nil(t, appErr)
	require.NotNil(t, db.inserted)
	assert.True(t, len(db.inserted.Password) > 0)
	assert.Equal(t, "PREFIX_pw", db.inserted.Password)
}

func TestUserService_CreateNewUser_DefaultHasherFallback(t *testing.T) {
	db := &mockDBUser{}
	// Manual struct init without hasher to ensure fallback logic
	svc := &UserService{Database: db, Logger: zap.NewNop(), WalletManager: nil}
	appErr := svc.CreateNewUser(&dto.UserRequest{Email: "a@b.com", Password: "pw"})
	require.Nil(t, appErr)
	require.NotNil(t, db.inserted)
	assert.NotEmpty(t, db.inserted.Password)
	assert.NotEqual(t, "pw", db.inserted.Password)
}
