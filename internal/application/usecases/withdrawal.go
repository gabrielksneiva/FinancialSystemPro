package usecases

import (
	"context"
	"errors"
	"fmt"

	"financial-system-pro/internal/domain/finance"
	"financial-system-pro/internal/infrastructure/eventbus"

	"github.com/google/uuid"
)

// WithdrawalUseCase handles withdrawals from user wallets to external fiat accounts.
// Flow: Wallet debited -> Ledger entry -> Fiat transfer initiated -> Event published
type WithdrawalUseCase struct {
	accounts      finance.AccountRepository
	wallets       finance.WalletRepository
	ledger        finance.LedgerRepository
	transactional finance.Transactional
	eventBus      eventbus.Bus
}

func NewWithdrawalUseCase(
	accounts finance.AccountRepository,
	wallets finance.WalletRepository,
	ledger finance.LedgerRepository,
	transactional finance.Transactional,
	eventBus eventbus.Bus,
) *WithdrawalUseCase {
	return &WithdrawalUseCase{
		accounts:      accounts,
		wallets:       wallets,
		ledger:        ledger,
		transactional: transactional,
		eventBus:      eventBus,
	}
}

type WithdrawalRequest struct {
	AccountID        string
	WalletID         string
	Amount           int64
	Currency         string
	Description      string
	DestinationID    string // External fiat account/bank reference
	WithdrawalMethod string // e.g., "bank_transfer", "pix", "wire"
}

type WithdrawalResult struct {
	WithdrawalID  string
	AccountID     string
	WalletID      string
	Amount        finance.Amount
	BalanceAfter  int64
	LedgerEntryID string
	Status        string // "pending", "processing", "completed", "failed"
}

func (uc *WithdrawalUseCase) Execute(ctx context.Context, req WithdrawalRequest) (*WithdrawalResult, error) {
	// Validate input
	if req.AccountID == "" {
		return nil, errors.New("account ID is required")
	}
	if req.WalletID == "" {
		return nil, errors.New("wallet ID is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}
	if req.DestinationID == "" {
		return nil, errors.New("destination ID is required")
	}

	amount, err := finance.NewAmount(req.Amount, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	var result *WithdrawalResult

	// Execute in transaction
	err = uc.transactional.WithinTx(ctx, func(txCtx context.Context) error {
		// Verify account exists and is active
		account, err := uc.accounts.Get(txCtx, req.AccountID)
		if err != nil {
			return fmt.Errorf("account not found: %w", err)
		}
		if account.Status() != finance.AccountStatusActive {
			return errors.New("account is not active")
		}

		// Get wallet
		wallet, err := uc.wallets.Get(txCtx, req.WalletID)
		if err != nil {
			return fmt.Errorf("wallet not found: %w", err)
		}
		if wallet.AccountID() != req.AccountID {
			return errors.New("wallet does not belong to account")
		}

		// Check sufficient balance
		if wallet.Balance().Value() < req.Amount {
			return fmt.Errorf("insufficient balance: have %d, need %d", 
				wallet.Balance().Value(), req.Amount)
		}

		// Debit wallet
		if err := wallet.Debit(amount); err != nil {
			return fmt.Errorf("failed to debit wallet: %w", err)
		}

		// Save wallet
		if err := uc.wallets.Save(txCtx, wallet); err != nil {
			return fmt.Errorf("failed to save wallet: %w", err)
		}

		// Create ledger entry
		ledgerID := uuid.New().String()
		description := req.Description
		if description == "" {
			description = fmt.Sprintf("Withdrawal to %s via %s", req.DestinationID, req.WithdrawalMethod)
		}

		entry, err := finance.NewLedgerEntry(
			ledgerID,
			wallet.ID(),
			finance.EntryTypeDebit,
			amount,
			wallet.Balance().Value(),
			description,
		)
		if err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		if err := uc.ledger.Append(txCtx, entry); err != nil {
			return fmt.Errorf("failed to append ledger entry: %w", err)
		}

		result = &WithdrawalResult{
			WithdrawalID:  uuid.New().String(),
			AccountID:     req.AccountID,
			WalletID:      req.WalletID,
			Amount:        amount,
			BalanceAfter:  wallet.Balance().Value(),
			LedgerEntryID: ledgerID,
			Status:        "pending", // Will be updated by fiat gateway handler
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event for async processing by fiat gateway
	event, _ := eventbus.NewBaseEvent(
		"withdrawal.initiated",
		req.AccountID,
		"account",
		map[string]interface{}{
			"withdrawal_id":     result.WithdrawalID,
			"account_id":        result.AccountID,
			"wallet_id":         result.WalletID,
			"amount":            result.Amount.Value(),
			"currency":          result.Amount.Currency(),
			"balance_after":     result.BalanceAfter,
			"ledger_entry_id":   result.LedgerEntryID,
			"destination_id":    req.DestinationID,
			"withdrawal_method": req.WithdrawalMethod,
		},
	)
	uc.eventBus.PublishAsync(ctx, event)

	return result, nil
}
