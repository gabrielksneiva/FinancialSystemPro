package services

import (
	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DatabasePort expõe apenas os métodos usados pelos services, permitindo mocking e testes.
type DatabasePort interface {
	FindUserByField(field string, value any) (*repositories.User, error)
	Insert(value any) error
	SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error
	GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error)
	Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error
	Balance(userID uuid.UUID) (decimal.Decimal, error)
	UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error
}

// NewDatabaseAdapter adapta *repositories.NewDatabase ao DatabasePort.
type NewDatabaseAdapter struct{ Inner *repositories.NewDatabase }

func (a *NewDatabaseAdapter) FindUserByField(field string, value any) (*repositories.User, error) {
	return a.Inner.FindUserByField(field, value)
}
func (a *NewDatabaseAdapter) Insert(value any) error { return a.Inner.Insert(value) }
func (a *NewDatabaseAdapter) SaveWalletInfo(userID uuid.UUID, tronAddress, encryptedPrivKey string) error {
	return a.Inner.SaveWalletInfo(userID, tronAddress, encryptedPrivKey)
}
func (a *NewDatabaseAdapter) GetWalletInfo(userID uuid.UUID) (*repositories.WalletInfo, error) {
	return a.Inner.GetWalletInfo(userID)
}
func (a *NewDatabaseAdapter) Transaction(userID uuid.UUID, amount decimal.Decimal, transactionType string) error {
	return a.Inner.Transaction(userID, amount, transactionType)
}
func (a *NewDatabaseAdapter) Balance(userID uuid.UUID) (decimal.Decimal, error) {
	return a.Inner.Balance(userID)
}
func (a *NewDatabaseAdapter) UpdateTransaction(txID uuid.UUID, updates map[string]interface{}) error {
	return a.Inner.UpdateTransaction(txID, updates)
}
