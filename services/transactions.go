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
	DB             *r.NewDatabase
	W              *w.TransactionWorkerPool
	TronWorkerPool *w.TronWorkerPool
	TronService    *TronService
	Logger         *zap.Logger
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

	// Se não foi especificado endereço de destino, usar a wallet automática do usuário
	destinationAddress := tronAddress
	if destinationAddress == "" {
		// Buscar wallet automática do usuário
		walletInfo, err := t.DB.GetWalletInfo(uid)
		if err != nil {
			t.Logger.Error("user has no auto-generated TRON wallet", zap.String("user_id", uid.String()), zap.Error(err))
			return nil, fmt.Errorf("TRON wallet not found. Please contact support to generate your wallet")
		}
		destinationAddress = walletInfo.TronAddress
		t.Logger.Info("using user's auto-generated wallet",
			zap.String("user_id", uid.String()),
			zap.String("wallet_address", destinationAddress),
		)
	} else {
		// Validar endereço TRON manualmente especificado
		if !t.TronService.ValidateAddress(destinationAddress) {
			return nil, fmt.Errorf("invalid TRON address: %s", destinationAddress)
		}
	}

	// Converter amount de decimal para SUN (1 TRX = 1.000.000 SUN)
	amountInSun := amount.Mul(decimal.NewFromInt(1000000)).BigInt().Int64()

	// Descriptografar private key da carteira
	walletPrivKey := ""
	if destinationAddress == tronAddress {
		// Se usando auto-wallet, descriptografar a private key
		walletInfo, err := t.DB.GetWalletInfo(uid)
		if err == nil && walletInfo != nil && walletInfo.EncryptedPrivKey != "" {
			decryptedKey, err := t.decryptPrivateKey(walletInfo.EncryptedPrivKey)
			if err != nil {
				t.Logger.Warn("error decrypting private key",
					zap.String("user_id", uid.String()),
					zap.Error(err),
				)
				// Continuar sem private key - será enviado manualmente depois
			} else {
				walletPrivKey = decryptedKey
			}
		}
	}

	t.Logger.Info("TRON withdraw request received",
		zap.String("user_id", uid.String()),
		zap.String("destination", destinationAddress),
		zap.String("amount_sun", fmt.Sprintf("%d", amountInSun)),
		zap.Bool("has_private_key", walletPrivKey != ""),
	)

	// Debitar do balance interno imediatamente
	err = t.DB.Transaction(uid, amount, "withdraw")
	if err != nil {
		t.Logger.Error("withdraw transaction failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	// Tentar enviar a transação TRON
	var txHash string
	var sendError error
	status := "pending_broadcast" // Status padrão: pendente de broadcast

	if walletPrivKey != "" && t.TronService != nil {
		// Obter endereço da carteira para usar como origem
		userWallet, errWallet := t.DB.GetWalletInfo(uid)
		if errWallet == nil && userWallet != nil {
			// Tentar enviar a transação
			txHash, sendError = t.TronService.SendTransaction(
				userWallet.TronAddress,
				destinationAddress,
				amountInSun,
				walletPrivKey,
			)

			if sendError != nil {
				t.Logger.Warn("error sending TRON transaction",
					zap.String("user_id", uid.String()),
					zap.String("from", userWallet.TronAddress),
					zap.String("to", destinationAddress),
					zap.Error(sendError),
				)
				status = "send_failed"
			} else {
				t.Logger.Info("TRON transaction sent successfully",
					zap.String("user_id", uid.String()),
					zap.String("tx_hash", txHash),
					zap.String("to", destinationAddress),
				)
				status = "confirmed" // Será atualizado após confirmação
			}
		}
	}

	// Criar registro de transação
	txRecord := &r.Transaction{
		AccountID:    uid,
		Amount:       amount,
		Type:         "withdraw",
		Category:     "debit",
		Description:  fmt.Sprintf("TRON withdraw to %s", destinationAddress),
		TronTxHash:   stringPtr(txHash), // Hash da transação se enviada com sucesso
		TronTxStatus: &status,
	}

	err = t.DB.Insert(txRecord)
	if err != nil {
		t.Logger.Error("withdraw record insertion failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	// Submeter job de confirmação TRON se a TX foi enviada com sucesso
	if txHash != "" && t.TronWorkerPool != nil && status == "confirmed" {
		t.TronWorkerPool.SubmitConfirmationJob(uid, txRecord.ID, txHash, callbackURL)
		t.Logger.Info("TX confirmation job submitted",
			zap.String("user_id", uid.String()),
			zap.String("tx_hash", txHash),
			zap.String("tx_id", txRecord.ID.String()),
		)
	}

	t.Logger.Info("TRON withdraw registered",
		zap.String("user_id", uid.String()),
		zap.String("amount", amount.String()),
		zap.String("to_address", destinationAddress),
		zap.String("tx_id", txRecord.ID.String()),
	)

	return &ServiceResponse{
		StatusCode: fiber.StatusAccepted,
		Body: fiber.Map{
			"message":     "Withdrawal registered and pending TRON blockchain broadcast",
			"tx_id":       txRecord.ID.String(),
			"amount":      amount.String(),
			"to_address":  destinationAddress,
			"status":      "pending_broadcast",
			"description": "Your withdrawal will be broadcast to TRON blockchain shortly",
		},
	}, nil
}

// decryptPrivateKey descriptografa a private key armazenada
func (t *NewTransactionService) decryptPrivateKey(encryptedKey string) (string, error) {
	if encryptedKey == "" {
		return "", fmt.Errorf("encrypted key is empty")
	}

	// Usar a função de utils para descriptografar
	return utils.DecryptPrivateKey(encryptedKey)
}

// stringPtr retorna um pointer para uma string
func stringPtr(s string) *string {
	return &s
}
