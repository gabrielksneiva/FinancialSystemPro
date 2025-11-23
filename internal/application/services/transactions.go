package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"financial-system-pro/internal/domain/errors"
	r "financial-system-pro/internal/infrastructure/database"
	w "financial-system-pro/internal/infrastructure/queue"
	"financial-system-pro/internal/shared/utils"
	"fmt"
	"net/http"
	"os"
	"time"

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
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
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
		return nil, errors.NewDatabaseError("deposit failed", nil)
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
		return nil, errors.NewDatabaseError("deposit record failed", nil)
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
// SEMPRE envia da carteira do cofre (vault) para a carteira TRON do usuário (buscada no banco)
func (t *NewTransactionService) WithdrawTron(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
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

	// Buscar a wallet TRON do usuário no banco de dados
	walletInfo, err := t.DB.GetWalletInfo(uid)
	if err != nil {
		t.Logger.Error("user has no auto-generated TRON wallet", zap.String("user_id", uid.String()), zap.Error(err))
		return nil, fmt.Errorf("TRON wallet not found. Please contact support to generate your wallet")
	}

	destinationAddress := walletInfo.TronAddress
	t.Logger.Info("withdraw to user's TRON wallet from database",
		zap.String("user_id", uid.String()),
		zap.String("wallet_address", destinationAddress),
	)

	// Verificar se o cofre TRON está configurado
	if !t.TronService.HasVaultConfigured() {
		t.Logger.Error("TRON vault not configured")
		return nil, fmt.Errorf("TRON vault is not configured. Please contact administrator")
	}

	vaultAddress := t.TronService.GetVaultAddress()
	vaultPrivateKey := t.TronService.GetVaultPrivateKey()

	// Converter amount de decimal para SUN (1 TRX = 1.000.000 SUN)
	amountInSun := amount.Mul(decimal.NewFromInt(1000000)).BigInt().Int64()

	t.Logger.Info("TRON withdraw request received",
		zap.String("user_id", uid.String()),
		zap.String("from_vault", vaultAddress),
		zap.String("to_user_wallet", destinationAddress),
		zap.String("amount_sun", fmt.Sprintf("%d", amountInSun)),
	)

	// Criar registro de transação com status 'pending'
	txRecord := &r.Transaction{
		AccountID:    uid,
		Amount:       amount,
		Type:         "withdraw",
		Category:     "debit",
		Description:  fmt.Sprintf("TRON withdraw to %s", destinationAddress),
		TronTxStatus: stringPtr("pending"),
	}

	err = t.DB.Insert(txRecord)
	if err != nil {
		t.Logger.Error("withdraw record insertion failed", zap.String("account_id", uid.String()), zap.Error(err))
		return nil, err
	}

	// Enviar webhook de status 'pending'
	if callbackURL != "" {
		t.sendWebhook(callbackURL, map[string]interface{}{
			"status":      "pending",
			"tx_id":       txRecord.ID.String(),
			"user_id":     uid.String(),
			"amount":      amount.String(),
			"to_address":  destinationAddress,
			"timestamp":   time.Now().Unix(),
			"description": "Withdraw request received and queued",
		})
	}

	// Debitar do balance interno imediatamente
	err = t.DB.Transaction(uid, amount, "withdraw")
	if err != nil {
		t.Logger.Error("withdraw transaction failed", zap.String("account_id", uid.String()), zap.Error(err))
		// Reverter status para failed
		t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": "failed"})
		return nil, err
	}

	// Atualizar status para 'broadcasting'
	err = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": "broadcasting"})
	if err != nil {
		t.Logger.Warn("failed to update status to broadcasting", zap.Error(err))
	}

	// Enviar webhook de status 'broadcasting'
	if callbackURL != "" {
		t.sendWebhook(callbackURL, map[string]interface{}{
			"status":      "broadcasting",
			"tx_id":       txRecord.ID.String(),
			"user_id":     uid.String(),
			"amount":      amount.String(),
			"to_address":  destinationAddress,
			"timestamp":   time.Now().Unix(),
			"description": "Broadcasting transaction to TRON network",
		})
	}

	// Enviar a transação TRON do cofre para a wallet do usuário
	var txHash string
	var sendError error
	status := "broadcasting"

	// Tentar enviar a transação do VAULT para o usuário
	txHash, sendError = t.TronService.SendTransaction(
		vaultAddress,
		destinationAddress,
		amountInSun,
		vaultPrivateKey,
	)

	if sendError != nil {
		t.Logger.Error("error sending TRON transaction from vault",
			zap.String("user_id", uid.String()),
			zap.String("from_vault", vaultAddress),
			zap.String("to_user", destinationAddress),
			zap.Error(sendError),
		)
		status = "failed"

		// Atualizar status no banco
		t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": status})

		// Enviar webhook de falha
		if callbackURL != "" {
			t.sendWebhook(callbackURL, map[string]interface{}{
				"status":      "failed",
				"tx_id":       txRecord.ID.String(),
				"user_id":     uid.String(),
				"amount":      amount.String(),
				"to_address":  destinationAddress,
				"timestamp":   time.Now().Unix(),
				"error":       sendError.Error(),
				"description": "Transaction broadcast failed",
			})
		}

		return nil, fmt.Errorf("failed to broadcast transaction: %w", sendError)
	}

	// Broadcast bem-sucedido
	t.Logger.Info("TRON transaction sent successfully from vault",
		zap.String("user_id", uid.String()),
		zap.String("tx_hash", txHash),
		zap.String("from", vaultAddress),
		zap.String("to", destinationAddress),
	)
	status = "broadcast_success"

	// Atualizar registro com hash e status
	err = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{
		"tron_tx_hash":   txHash,
		"tron_tx_status": status,
	})
	if err != nil {
		t.Logger.Error("failed to update tx hash", zap.Error(err))
	}

	// Enviar webhook de broadcast_success
	if callbackURL != "" {
		t.sendWebhook(callbackURL, map[string]interface{}{
			"status":       "broadcast_success",
			"tx_id":        txRecord.ID.String(),
			"tx_hash":      txHash,
			"user_id":      uid.String(),
			"amount":       amount.String(),
			"to_address":   destinationAddress,
			"from_address": vaultAddress,
			"timestamp":    time.Now().Unix(),
			"description":  "Transaction successfully broadcast to TRON network",
			"explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", txHash),
		})
	}

	// Submeter job de confirmação TRON
	if t.TronWorkerPool != nil {
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
			"message":      "Withdrawal broadcast to TRON blockchain",
			"tx_id":        txRecord.ID.String(),
			"tx_hash":      txHash,
			"amount":       amount.String(),
			"to_address":   destinationAddress,
			"status":       status,
			"description":  "Transaction is awaiting confirmations on TRON network",
			"explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", txHash),
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

// sendWebhook envia webhook com retry automático e assinatura HMAC
func (t *NewTransactionService) sendWebhook(url string, data map[string]interface{}) {
	if url == "" {
		return
	}

	// Adicionar metadados
	data["event_type"] = "tron_transaction_update"
	data["webhook_version"] = "1.0"

	payload, err := json.Marshal(data)
	if err != nil {
		t.Logger.Error("failed to marshal webhook payload", zap.Error(err))
		return
	}

	// Tentar enviar até 3 vezes
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			t.Logger.Error("failed to create webhook request", zap.Error(err))
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "FinancialSystemPro/1.0")
		req.Header.Set("X-Webhook-Event", "tron_transaction_update")
		req.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", attempt))

		// Adicionar assinatura HMAC se tiver secret key configurada
		if secret := os.Getenv("WEBHOOK_SECRET"); secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(payload)
			signature := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-Webhook-Signature", signature)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Logger.Warn("webhook request failed",
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.Logger.Info("webhook sent successfully",
				zap.String("url", url),
				zap.Int("status_code", resp.StatusCode),
				zap.String("event", data["status"].(string)),
			)
			return
		}

		t.Logger.Warn("webhook returned non-success status",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.Int("attempt", attempt),
		)

		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
}

// GetWalletInfo retorna informações da wallet do usuário
func (t *NewTransactionService) GetWalletInfo(userID uuid.UUID) (*r.WalletInfo, error) {
	return t.DB.GetWalletInfo(userID)
}
