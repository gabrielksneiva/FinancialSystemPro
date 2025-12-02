package usecases

import (
	"context"
	"errors"
	"fmt"

	"financial-system-pro/internal/domain/finance"
	"financial-system-pro/internal/infrastructure/eventbus"

	"github.com/google/uuid"
)

// DepositUseCase handles fiat deposits into user wallets.
// Flow: Fiat received -> Wallet credited -> Ledger entry -> Event published
type DepositUseCase struct {
	accounts      finance.AccountRepository
	wallets       finance.WalletRepository
	ledger        finance.LedgerRepository
	transactional finance.Transactional
	eventBus      eventbus.Bus
}

func NewDepositUseCase(
	accounts finance.AccountRepository,
	wallets finance.WalletRepository,
	ledger finance.LedgerRepository,
	transactional finance.Transactional,
	eventBus eventbus.Bus,
) *DepositUseCase {
	return &DepositUseCase{
		accounts:      accounts,
		wallets:       wallets,
		ledger:        ledger,
		transactional: transactional,
		eventBus:      eventBus,
	}
}

type DepositRequest struct {
	AccountID   string
	WalletID    string
	Amount      int64
	Currency    string
	Description string
	ExternalID  string // Reference to fiat gateway transaction
}

type DepositResult struct {
	DepositID     string
	AccountID     string
	WalletID      string
	Amount        finance.Amount
	BalanceAfter  int64
	LedgerEntryID string
}

func (uc *DepositUseCase) Execute(ctx context.Context, req DepositRequest) (*DepositResult, error) {
	// Validate input
	if req.AccountID == "" {
		return nil, errors.New("account ID is required")
	}
	if req.WalletID == "" {
		return nil, errors.New("wallet ID is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("deposit amount must be positive")
	}

	amount, err := finance.NewAmount(req.Amount, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	var result *DepositResult

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

		// Credit wallet
		if err := wallet.Credit(amount); err != nil {
			return fmt.Errorf("failed to credit wallet: %w", err)
		}

		// Save wallet
		if err := uc.wallets.Save(txCtx, wallet); err != nil {
			return fmt.Errorf("failed to save wallet: %w", err)
		}

		// Create ledger entry
		ledgerID := uuid.New().String()
		description := req.Description
		if description == "" {
			description = fmt.Sprintf("Deposit from external ID: %s", req.ExternalID)
		}

		entry, err := finance.NewLedgerEntry(
			ledgerID,
			wallet.ID(),
			finance.EntryTypeCredit,
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

		result = &DepositResult{
			DepositID:     uuid.New().String(),
			AccountID:     req.AccountID,
			WalletID:      req.WalletID,
			Amount:        amount,
			BalanceAfter:  wallet.Balance().Value(),
			LedgerEntryID: ledgerID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event (outside transaction for async processing)
	event, _ := eventbus.NewBaseEvent(
		"deposit.completed",
		req.AccountID,
		"account",
		map[string]interface{}{
			"deposit_id":       result.DepositID,
			"account_id":       result.AccountID,
			"wallet_id":        result.WalletID,
			"amount":           result.Amount.Value(),
			"currency":         result.Amount.Currency(),
			"balance_after":    result.BalanceAfter,
			"ledger_entry_id":  result.LedgerEntryID,
			"external_id":      req.ExternalID,
		},
	)
	uc.eventBus.PublishAsync(ctx, event)

	return result, nil
}
