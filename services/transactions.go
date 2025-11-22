package services

import (
	"financial-system-pro/domain"
	r "financial-system-pro/repositories"
	"financial-system-pro/utils"
	w "financial-system-pro/workers"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type NewTransactionService struct {
	DB          *r.NewDatabase
	W           *w.TransactionWorkerPool
	TronService *TronService
	Logger      *zap.Logger
}

func (t *NewTransactionService) Deposit(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	idLocal := c.Locals("user_id")
	if idLocal == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}
	id, ok := idLocal.(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id format in context")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		t.Logger.Error("invalid user id format", zap.String("id", id), zap.Error(err))
		return nil, domain.NewValidationError("user_id", "Invalid user ID format")
	}

	if t.W != nil {
		job := w.TransactionJob{
			Type:        w.JobDeposit,
			Account:     uid,
			Amount:      amount,
			CallbackURL: callbackURL,
			JobID:       uuid.New(),
		}

		t.W.Jobs <- job
		t.Logger.Info("deposit job queued",
			zap.String("account_id", uid.String()),
			zap.String("amount", amount.String()),
			zap.String("job_id", job.JobID.String()),
		)

		return &ServiceResponse{
			StatusCode: fiber.StatusAccepted,
			Body: fiber.Map{
				"job_id": job.JobID.String(),
				"status": "pending",
			},
		}, nil
	}

	if err := t.DB.Transaction(uid, amount, "deposit"); err != nil {
		t.Logger.Error("deposit transaction failed",
			zap.String("account_id", uid.String()),
			zap.String("amount", amount.String()),
			zap.Error(err),
		)
		return nil, domain.NewDatabaseError("deposit failed", nil)
	}
	if err := t.DB.Insert(&r.Transaction{
		AccountID:   uid,
		Amount:      amount,
		Type:        "deposit",
		Category:    "credit",
		Description: "User deposit",
	}); err != nil {
		t.Logger.Error("deposit record insertion failed",
			zap.String("account_id", uid.String()),
			zap.Error(err),
		)
		return nil, domain.NewDatabaseError("deposit record failed", nil)
	}

	t.Logger.Info("deposit successful",
		zap.String("account_id", uid.String()),
		zap.String("amount", amount.String()),
	)
	return &ServiceResponse{
		StatusCode: fiber.StatusOK,
		Body:       fiber.Map{"message": "Deposit successfully"},
	}, nil
}

func (t *NewTransactionService) GetBalance(c *fiber.Ctx, userID string) (decimal.Decimal, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		t.Logger.Warn("invalid user id for balance check", zap.String("id", userID))
		return decimal.Zero, err
	}

	response, err := t.DB.Balance(uid)
	if err != nil {
		t.Logger.Error("balance query failed", zap.String("account_id", uid.String()), zap.Error(err))
		return decimal.Zero, err
	}

	t.Logger.Debug("balance retrieved", zap.String("account_id", uid.String()), zap.String("balance", response.String()))
	return response, nil
}

func (t *NewTransactionService) Withdraw(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	idLocal := c.Locals("user_id")
	if idLocal == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}
	id, ok := idLocal.(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id format in context")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		t.Logger.Error("invalid user id for withdraw", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	if t.W != nil {
		job := w.TransactionJob{
			Type:        w.JobWithdraw,
			Account:     uid,
			Amount:      amount,
			CallbackURL: callbackURL,
			JobID:       uuid.New(),
		}

		t.W.Jobs <- job
		t.Logger.Info("withdraw job queued", zap.String("account_id", uid.String()), zap.String("amount", amount.String()))
		return &ServiceResponse{
			StatusCode: fiber.StatusAccepted,
			Body: fiber.Map{
				"job_id": job.JobID.String(),
				"status": "pending",
			},
		}, nil
	}

	err = t.DB.Transaction(uid, amount, "withdraw")
	if err != nil {
		t.Logger.Error("withdraw transaction failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	err = t.DB.Insert(&r.Transaction{
		AccountID:   uid,
		Amount:      amount,
		Type:        "withdraw",
		Category:    "debit",
		Description: "User withdraw",
	})
	if err != nil {
		t.Logger.Error("withdraw record insertion failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	t.Logger.Info("withdraw successful", zap.String("account_id", uid.String()), zap.String("amount", amount.String()))
	return &ServiceResponse{
		StatusCode: fiber.StatusOK,
		Body:       fiber.Map{"message": "Withdraw successfully"},
	}, nil
}

func (t *NewTransactionService) Transfer(c *fiber.Ctx, amount decimal.Decimal, userTo, callbackURL string) (*ServiceResponse, error) {
	idLocal := c.Locals("user_id")
	if idLocal == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}
	id, ok := idLocal.(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id format in context")
	}
	userFrom, err := uuid.Parse(id)
	if err != nil {
		t.Logger.Error("invalid user id for transfer", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	if t.W != nil {
		job := w.TransactionJob{
			Type:        w.JobTransfer,
			Account:     userFrom,
			Amount:      amount,
			ToEmail:     userTo,
			CallbackURL: callbackURL,
			JobID:       uuid.New(),
		}

		t.W.Jobs <- job
		t.Logger.Info("transfer job queued",
			zap.String("from_id", userFrom.String()),
			zap.String("to_email", userTo),
			zap.String("amount", amount.String()),
		)

		return &ServiceResponse{
			StatusCode: fiber.StatusAccepted,
			Body: fiber.Map{
				"job_id": job.JobID.String(),
				"status": "pending",
			},
		}, nil
	}

	// fallback synchronous processing if worker pool is not initialized
	foundUser, err := t.DB.FindUserByField("email", userTo)
	if err != nil {
		t.Logger.Warn("transfer recipient not found", zap.String("email", userTo), zap.Error(err))
		return nil, err
	}
	destinyUserID := foundUser.ID

	foundUserFrom, err := t.DB.FindUserByField("id", userFrom.String())
	if err != nil {
		t.Logger.Error("transfer source user not found", zap.String("id", userFrom.String()), zap.Error(err))
		return nil, err
	}

	if err := t.DB.Transaction(userFrom, amount, "withdraw"); err != nil {
		t.Logger.Error("transfer withdraw failed", zap.String("from_id", userFrom.String()), zap.Error(err))
		return nil, err
	}
	if err := t.DB.Transaction(destinyUserID, amount, "deposit"); err != nil {
		t.Logger.Error("transfer deposit failed", zap.String("to_id", destinyUserID.String()), zap.Error(err))
		return nil, err
	}

	if err := t.DB.Insert(&r.Transaction{
		AccountID:   userFrom,
		Amount:      amount,
		Type:        "transfer",
		Category:    "debit",
		Description: "User transfer to " + userTo,
	}); err != nil {
		t.Logger.Error("transfer record insertion failed", zap.String("from_id", userFrom.String()), zap.Error(err))
		return nil, err
	}

	if err := t.DB.Insert(&r.Transaction{
		AccountID:   destinyUserID,
		Amount:      amount,
		Type:        "transfer",
		Category:    "credit",
		Description: "User transfer from " + foundUserFrom.Email,
	}); err != nil {
		t.Logger.Error("transfer record insertion failed", zap.String("to_id", destinyUserID.String()), zap.Error(err))
		return nil, err
	}

	t.Logger.Info("transfer successful",
		zap.String("from_id", userFrom.String()),
		zap.String("to_email", userTo),
		zap.String("amount", amount.String()),
	)
	return &ServiceResponse{
		StatusCode: fiber.StatusOK,
		Body:       fiber.Map{"message": "Transfer successfully"},
	}, nil
}

// WithdrawTron realiza um saque direto para TRON blockchain
func (t *NewTransactionService) WithdrawTron(c *fiber.Ctx, amount decimal.Decimal, tronAddress, callbackURL string) (*ServiceResponse, error) {
	idLocal := c.Locals("user_id")
	if idLocal == nil {
		return nil, fmt.Errorf("user_id not found in context")
	}
	id, ok := idLocal.(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id format in context")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		t.Logger.Error("invalid user id for tron withdraw", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	// Validar endereço TRON
	if !t.TronService.ValidateAddress(tronAddress) {
		return nil, fmt.Errorf("invalid TRON address: %s", tronAddress)
	}

	// Verificar balance disponível
	balance, err := t.DB.Balance(uid)
	if err != nil {
		t.Logger.Error("error fetching balance for TRON withdraw", zap.String("user_id", uid.String()), zap.Error(err))
		return nil, err
	}

	if balance.LessThan(amount) {
		return nil, fmt.Errorf("insufficient balance: have %s, need %s", balance.String(), amount.String())
	}

	// Verificar se usuário tem carteira TRON registrada
	walletInfo, err := t.DB.GetWalletInfo(uid)
	if err != nil {
		t.Logger.Error("user has no TRON wallet registered", zap.String("user_id", uid.String()), zap.Error(err))
		return nil, fmt.Errorf("TRON wallet not configured. Please setup your wallet first")
	}

	// Descriptografar private key
	privKey, err := utils.DecryptPrivateKey(walletInfo.EncryptedPrivKey)
	if err != nil {
		t.Logger.Error("error decrypting private key", zap.String("user_id", uid.String()), zap.Error(err))
		return nil, fmt.Errorf("error decrypting wallet: %v", err)
	}

	// Converter amount de decimal para SUN (1 TRX = 1.000.000 SUN)
	amountInSun := amount.Mul(decimal.NewFromInt(1000000)).BigInt().Int64()

	// Enviar transação TRON
	txHash, err := t.TronService.SendTransaction(walletInfo.TronAddress, tronAddress, amountInSun, privKey)
	if err != nil {
		t.Logger.Error("error sending TRON transaction",
			zap.String("user_id", uid.String()),
			zap.String("to_address", tronAddress),
			zap.Error(err))
		return nil, fmt.Errorf("error sending transaction: %v", err)
	}

	// Debitar do balance interno
	err = t.DB.Transaction(uid, amount, "withdraw")
	if err != nil {
		t.Logger.Error("withdraw transaction failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	// Criar registro de transação com hash TRON
	status := "pending"
	txRecord := &r.Transaction{
		AccountID:    uid,
		Amount:       amount,
		Type:         "withdraw",
		Category:     "debit",
		Description:  fmt.Sprintf("TRON withdraw to %s", tronAddress),
		TronTxHash:   &txHash,
		TronTxStatus: &status,
	}

	err = t.DB.Insert(txRecord)
	if err != nil {
		t.Logger.Error("withdraw record insertion failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	t.Logger.Info("TRON withdraw transaction submitted",
		zap.String("user_id", uid.String()),
		zap.String("amount", amount.String()),
		zap.String("to_address", tronAddress),
		zap.String("tx_hash", txHash),
	)

	return &ServiceResponse{
		StatusCode: fiber.StatusAccepted,
		Body: fiber.Map{
			"message":     "Withdrawal transaction submitted to TRON blockchain",
			"tx_hash":     txHash,
			"amount":      amount.String(),
			"to_address":  tronAddress,
			"status":      "pending",
			"description": "Transaction is pending TRON blockchain confirmation",
		},
	}, nil
}
