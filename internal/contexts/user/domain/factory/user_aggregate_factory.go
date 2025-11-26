package factory

import (
	"errors"
	"financial-system-pro/internal/contexts/user/domain/entity"
	"financial-system-pro/internal/contexts/user/domain/valueobject"
)

// UserAggregateFactory handles complex construction of UserAggregate
type UserAggregateFactory struct{}

// NewUserAggregateFactory creates a new factory instance
func NewUserAggregateFactory() *UserAggregateFactory {
	return &UserAggregateFactory{}
}

// Create creates a UserAggregate with full validation
func (f *UserAggregateFactory) Create(
	emailStr string,
	passwordStr string,
	address string,
	encryptedPrivKey string,
) (*entity.UserAggregate, error) {
	// Validate and create value objects
	email, err := valueobject.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}

	password, err := valueobject.HashFromRaw(passwordStr)
	if err != nil {
		return nil, err
	}

	// Create aggregate
	aggregate, err := entity.NewUserAggregate(email, password, address, encryptedPrivKey)
	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

// CreateSimple creates a UserAggregate without blockchain wallet (for testing/internal)
func (f *UserAggregateFactory) CreateSimple(
	emailStr string,
	passwordStr string,
) (*entity.UserAggregate, error) {
	return f.Create(emailStr, passwordStr, "", "")
}

// CreateWithWallet creates a UserAggregate with blockchain wallet address
func (f *UserAggregateFactory) CreateWithWallet(
	emailStr string,
	passwordStr string,
	walletAddress string,
	encryptedPrivateKey string,
) (*entity.UserAggregate, error) {
	if walletAddress == "" {
		return nil, errors.New("wallet address cannot be empty when creating with wallet")
	}

	return f.Create(emailStr, passwordStr, walletAddress, encryptedPrivateKey)
}
