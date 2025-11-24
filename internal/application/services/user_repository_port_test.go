package services

import (
	"context"
	infra "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// fakeDatabasePort minimal para testar adapter sem depender do DB real
type fakeDatabasePort struct{ users map[uuid.UUID]*infra.User }

func (f *fakeDatabasePort) FindUserByField(field string, value any) (*infra.User, error) {
	switch field {
	case "email":
		for _, u := range f.users {
			if u.Email == value {
				return u, nil
			}
		}
	case "id":
		idStr, _ := value.(string)
		for _, u := range f.users {
			if u.ID.String() == idStr {
				return u, nil
			}
		}
	}
	return nil, nil
}
func (f *fakeDatabasePort) Insert(v any) error {
	if u, ok := v.(*infra.User); ok {
		f.users[u.ID] = u
	}
	return nil
}
func (f *fakeDatabasePort) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return nil
}
func (f *fakeDatabasePort) GetWalletInfo(userID uuid.UUID) (*infra.WalletInfo, error) {
	return nil, nil
}
func (f *fakeDatabasePort) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return nil
}
func (f *fakeDatabasePort) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (f *fakeDatabasePort) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	return nil
}

// Adaptar tipos necessários (mock simplificado para decimal)

func TestUserRepositoryAdapter(t *testing.T) {
	db := &fakeDatabasePort{users: make(map[uuid.UUID]*infra.User)}
	adapter := NewUserRepositoryAdapter(db)
	ctx := context.Background()

	u := &infra.User{ID: uuid.New(), Email: "repo@test.com", Password: "x"}
	require.NoError(t, adapter.Save(ctx, u))

	foundByEmail, err := adapter.FindByEmail(ctx, "repo@test.com")
	require.NoError(t, err)
	require.NotNil(t, foundByEmail)
	require.Equal(t, u.ID, foundByEmail.ID)

	foundByID, err := adapter.FindByID(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, foundByID)
	require.Equal(t, "repo@test.com", foundByID.Email)
}
