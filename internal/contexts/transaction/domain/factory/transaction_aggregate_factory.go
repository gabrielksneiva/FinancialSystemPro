package factory

import (
	"errors"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	domainEntities "financial-system-pro/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionAggregateFactory handles complex construction of TransactionAggregate
type TransactionAggregateFactory struct{}

// NewTransactionAggregateFactory creates a new factory instance
func NewTransactionAggregateFactory() *TransactionAggregateFactory {
	return &TransactionAggregateFactory{}
}

// CreateDeposit creates a TransactionAggregate for a deposit operation
func (f *TransactionAggregateFactory) CreateDeposit(
	userID uuid.UUID,
	amount decimal.Decimal,
	currency string,
	blockchain domainEntities.BlockchainType,
	walletAddress string,
) (*entity.TransactionAggregate, error) {
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	if amount.IsNegative() || amount.IsZero() {
		return nil, errors.New("deposit amount must be positive")
	}

	if currency == "" {
		currency = "USD"
	}

	if blockchain == "" {
		return nil, errors.New("blockchain type is required for deposits")
	}

	if walletAddress == "" {
		return nil, errors.New("wallet address is required for deposits")
	}

	aggregate, err := entity.NewTransactionAggregate(
		userID,
		entity.TransactionTypeDeposit,
		amount,
		currency,
		string(blockchain),
		walletAddress,
	)

	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

// CreateWithdrawal creates a TransactionAggregate for a withdrawal operation
func (f *TransactionAggregateFactory) CreateWithdrawal(
	userID uuid.UUID,
	amount decimal.Decimal,
	currency string,
	blockchain domainEntities.BlockchainType,
	toAddress string,
) (*entity.TransactionAggregate, error) {
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	if amount.IsNegative() || amount.IsZero() {
		return nil, errors.New("withdrawal amount must be positive")
	}

	if currency == "" {
		currency = "USD"
	}

	if blockchain == "" {
		return nil, errors.New("blockchain type is required for withdrawals")
	}

	if toAddress == "" {
		return nil, errors.New("destination address is required for withdrawals")
	}

	aggregate, err := entity.NewTransactionAggregate(
		userID,
		entity.TransactionTypeWithdraw,
		amount,
		currency,
		string(blockchain),
		toAddress,
	)

	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

// CreateTransfer creates a TransactionAggregate for an internal transfer
func (f *TransactionAggregateFactory) CreateTransfer(
	userID uuid.UUID,
	amount decimal.Decimal,
	currency string,
	toUserID uuid.UUID,
	reference string,
) (*entity.TransactionAggregate, error) {
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	if toUserID == uuid.Nil {
		return nil, errors.New("destination userID cannot be empty")
	}

	if userID == toUserID {
		return nil, errors.New("cannot transfer to the same user")
	}

	if amount.IsNegative() || amount.IsZero() {
		return nil, errors.New("transfer amount must be positive")
	}

	if currency == "" {
		currency = "USD"
	}

	aggregate, err := entity.NewTransactionAggregate(
		userID,
		entity.TransactionTypeTransfer,
		amount,
		currency,
		"INTERNAL",
		reference,
	)

	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

// CreateBlockchainTransaction creates a generic blockchain transaction
func (f *TransactionAggregateFactory) CreateBlockchainTransaction(
	userID uuid.UUID,
	txType entity.TransactionType,
	amount decimal.Decimal,
	currency string,
	blockchain domainEntities.BlockchainType,
	address string,
) (*entity.TransactionAggregate, error) {
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be empty")
	}

	if amount.IsNegative() || amount.IsZero() {
		return nil, errors.New("amount must be positive")
	}

	if currency == "" {
		currency = "USD"
	}

	if blockchain == "" {
		return nil, errors.New("blockchain type is required")
	}

	if address == "" {
		return nil, errors.New("blockchain address is required")
	}

	// Validate transaction type
	validTypes := map[entity.TransactionType]bool{
		entity.TransactionTypeDeposit:  true,
		entity.TransactionTypeWithdraw: true,
	}

	if !validTypes[txType] {
		return nil, errors.New("invalid transaction type for blockchain operation")
	}

	aggregate, err := entity.NewTransactionAggregate(
		userID,
		txType,
		amount,
		currency,
		string(blockchain),
		address,
	)

	if err != nil {
		return nil, err
	}

	return aggregate, nil
}
