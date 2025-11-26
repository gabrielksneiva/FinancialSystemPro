package services

import (
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"
)

// fake underlying database implementing called methods
// fakeNewDatabase mimics the subset of methods used by adapter
// type fakeNewDatabase struct{}

//	func (f *fakeNewDatabase) FindUserByField(field string, value any) (*repositories.User, error) {
//		return &repositories.User{Email: "a@b.com"}, nil
//	}
//
// func (f *fakeNewDatabase) Insert(value any) error { return nil }
//
//	func (f *fakeNewDatabase) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
//		return nil
//	}
//
//	func (f *fakeNewDatabase) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
//		return &repositories.WalletInfo{TronAddress: "ADDR"}, nil
//	}
//
//	func (f *fakeNewDatabase) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
//		return nil
//	}
//
//	func (f *fakeNewDatabase) Balance(userID uuid.UUID) (decimal.Decimal, error) {
//		return decimal.NewFromInt(42), nil
//	}
// func (f *fakeNewDatabase) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
// 	return nil
// }

func TestNewDatabaseAdapter_Structure(t *testing.T) {
	// Provide zero-value repositories.NewDatabase with nil DB, only structural test
	adapter := &NewDatabaseAdapter{Inner: &repositories.NewDatabase{}}
	if adapter.Inner == nil {
		t.Fatalf("inner not set")
	}
}
