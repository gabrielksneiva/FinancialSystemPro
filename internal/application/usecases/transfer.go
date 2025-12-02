package usecases

import (
	"context"
	"errors"
	"fmt"

	"financial-system-pro/internal/domain/finance"
	"financial-system-pro/internal/infrastructure/eventbus"

	"github.com/google/uuid"
)

// TransferUseCase handles internal transfers between wallets.
// Flow: Source debited -> Destination credited -> Two ledger entries -> Event published
type TransferUseCase struct {
	accounts      finance.AccountRepository
	wallets       finance.WalletRepository
	ledger        finance.LedgerRepository
	transactional finance.Transactional
	eventBus      eventbus.Bus
}

func NewTransferUseCase(
	accounts finance.AccountRepository,
	wallets finance.WalletRepository,
	ledger finance.LedgerRepository,
	transactional finance.Transactional,
	eventBus eventbus.Bus,
) *TransferUseCase {
	return &TransferUseCase{
		accounts:      accounts,
		wallets:       wallets,
		ledger:        ledger,
		transactional: transactional,
		eventBus:      eventBus,
	}
}

type TransferRequest struct {
	FromAccountID string
	FromWalletID  string
	ToAccountID   string
	ToWalletID    string
	Amount        int64
	Currency      string
	Description   string
}

type TransferResult struct {
	TransferID           string
	FromWalletID         string
	ToWalletID           string
	Amount               finance.Amount
	FromBalanceAfter     int64
	ToBalanceAfter       int64
	FromLedgerEntryID    string
	ToLedgerEntryID      string
}

func (uc *TransferUseCase) Execute(ctx context.Context, req TransferRequest) (*TransferResult, error) {
	// Validate input
	if req.FromAccountID == "" || req.FromWalletID == "" {
		return nil, errors.New("source account and wallet are required")
	}
	if req.ToAccountID == "" || req.ToWalletID == "" {
		return nil, errors.New("destination account and wallet are required")
	}
	if req.FromWalletID == req.ToWalletID {
		return nil, errors.New("cannot transfer to same wallet")
	}
	if req.Amount <= 0 {
		return nil, errors.New("transfer amount must be positive")
	}

	amount, err := finance.NewAmount(req.Amount, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	var result *TransferResult

	// Execute in transaction for atomicity
	err = uc.transactional.WithinTx(ctx, func(txCtx context.Context) error {
		// Verify source account
		fromAccount, err := uc.accounts.Get(txCtx, req.FromAccountID)
		if err != nil {
			return fmt.Errorf("source account not found: %w", err)
		}
		if fromAccount.Status() != finance.AccountStatusActive {
			return errors.New("source account is not active")
		}

		// Verify destination account
		toAccount, err := uc.accounts.Get(txCtx, req.ToAccountID)
		if err != nil {
			return fmt.Errorf("destination account not found: %w", err)
		}
		if toAccount.Status() != finance.AccountStatusActive {
			return errors.New("destination account is not active")
		}

		// Get source wallet
		fromWallet, err := uc.wallets.Get(txCtx, req.FromWalletID)
		if err != nil {
			return fmt.Errorf("source wallet not found: %w", err)
		}
		if fromWallet.AccountID() != req.FromAccountID {
			return errors.New("source wallet does not belong to source account")
		}

		// Get destination wallet
		toWallet, err := uc.wallets.Get(txCtx, req.ToWalletID)
		if err != nil {
			return fmt.Errorf("destination wallet not found: %w", err)
		}
		if toWallet.AccountID() != req.ToAccountID {
			return errors.New("destination wallet does not belong to destination account")
		}

		// Check currency match
		if fromWallet.Balance().Currency() != toWallet.Balance().Currency() {
			return fmt.Errorf("currency mismatch: %s vs %s", 
				fromWallet.Balance().Currency(), toWallet.Balance().Currency())
		}

		// Check sufficient balance
		if fromWallet.Balance().Value() < req.Amount {
			return fmt.Errorf("insufficient balance: have %d, need %d",
				fromWallet.Balance().Value(), req.Amount)
		}

		// Debit source
		if err := fromWallet.Debit(amount); err != nil {
			return fmt.Errorf("failed to debit source wallet: %w", err)
		}

		// Credit destination
		if err := toWallet.Credit(amount); err != nil {
			return fmt.Errorf("failed to credit destination wallet: %w", err)
		}

		// Save wallets
		if err := uc.wallets.Save(txCtx, fromWallet); err != nil {
			return fmt.Errorf("failed to save source wallet: %w", err)
		}
		if err := uc.wallets.Save(txCtx, toWallet); err != nil {
			return fmt.Errorf("failed to save destination wallet: %w", err)
		}

		// Create ledger entries
		transferID := uuid.New().String()
		description := req.Description
		if description == "" {
			description = fmt.Sprintf("Transfer %s", transferID)
		}

		fromLedgerID := uuid.New().String()
		fromEntry, err := finance.NewLedgerEntry(
			fromLedgerID,
			fromWallet.ID(),
			finance.EntryTypeDebit,
			amount,
			fromWallet.Balance().Value(),
			fmt.Sprintf("%s (to %s)", description, req.ToWalletID),
		)
		if err != nil {
			return fmt.Errorf("failed to create source ledger entry: %w", err)
		}

		toLedgerID := uuid.New().String()
		toEntry, err := finance.NewLedgerEntry(
			toLedgerID,
			toWallet.ID(),
			finance.EntryTypeCredit,
			amount,
			toWallet.Balance().Value(),
			fmt.Sprintf("%s (from %s)", description, req.FromWalletID),
		)
		if err != nil {
			return fmt.Errorf("failed to create destination ledger entry: %w", err)
		}

		if err := uc.ledger.Append(txCtx, fromEntry); err != nil {
			return fmt.Errorf("failed to append source ledger entry: %w", err)
		}
		if err := uc.ledger.Append(txCtx, toEntry); err != nil {
			return fmt.Errorf("failed to append destination ledger entry: %w", err)
		}

		result = &TransferResult{
			TransferID:        transferID,
			FromWalletID:      req.FromWalletID,
			ToWalletID:        req.ToWalletID,
			Amount:            amount,
			FromBalanceAfter:  fromWallet.Balance().Value(),
			ToBalanceAfter:    toWallet.Balance().Value(),
			FromLedgerEntryID: fromLedgerID,
			ToLedgerEntryID:   toLedgerID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event
	event, _ := eventbus.NewBaseEvent(
		"transfer.completed",
		req.FromAccountID,
		"account",
		map[string]interface{}{
			"transfer_id":          result.TransferID,
			"from_account_id":      req.FromAccountID,
			"from_wallet_id":       result.FromWalletID,
			"to_account_id":        req.ToAccountID,
			"to_wallet_id":         result.ToWalletID,
			"amount":               result.Amount.Value(),
			"currency":             result.Amount.Currency(),
			"from_balance_after":   result.FromBalanceAfter,
			"to_balance_after":     result.ToBalanceAfter,
			"from_ledger_entry_id": result.FromLedgerEntryID,
			"to_ledger_entry_id":   result.ToLedgerEntryID,
		},
	)
	uc.eventBus.PublishAsync(ctx, event)

	return result, nil
}
