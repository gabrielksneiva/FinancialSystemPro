package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"financial-system-pro/internal/domain/errors"
	r "financial-system-pro/internal/infrastructure/database"
	w "financial-system-pro/internal/infrastructure/queue"
	"financial-system-pro/internal/shared/events"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type TransactionService struct {
	DB          DatabasePort
	Queue       QueuePort
	Tron        TronPort
	TronConfirm TronConfirmationPort
	Events      EventsPort
	Logger      *zap.Logger
}

func NewTransactionService(db DatabasePort, queue QueuePort, tron TronPort, tronConfirm TronConfirmationPort, events EventsPort, logger *zap.Logger) *TransactionService {
	return &TransactionService{DB: db, Queue: queue, Tron: tron, TronConfirm: tronConfirm, Events: events, Logger: logger}
}

func (t *TransactionService) Deposit(userID string, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
	}

	job := w.TransactionJob{Type: w.JobDeposit, Account: uid, Amount: amount, CallbackURL: callbackURL, JobID: uuid.New()}
	if t.Queue != nil {
		_ = t.Queue.QueueTransaction(job)
		return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"job_id": job.JobID.String(), "status": "pending"}}, nil
	}

	if err := t.DB.Transaction(uid, amount, "deposit"); err != nil {
		if t.Events != nil {
			t.Events.PublishAsync(context.Background(), events.NewTransactionFailedEvent(uid, "deposit", amount, err.Error(), "DEPOSIT_FAILED"))
		}
		return nil, errors.NewDatabaseError("deposit failed", nil)
	}
	if err := t.DB.Insert(&r.Transaction{AccountID: uid, Amount: amount, Type: "deposit", Category: "credit", Description: "User deposit"}); err != nil {
		return nil, errors.NewDatabaseError("deposit record failed", nil)
	}

	if t.Events != nil {
		t.Events.PublishAsync(context.Background(), events.NewDepositCompletedEvent(uid, amount, fmt.Sprintf("dep_%s", uuid.New().String()[:8])))
	}
	return &ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"message": "Deposit successfully"}}, nil
}

func (t *TransactionService) GetBalance(userID string) (decimal.Decimal, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return decimal.Zero, err
	}
	response, err := t.DB.Balance(uid)
	if err != nil {
		return decimal.Zero, err
	}
	return response, nil
}

func (t *TransactionService) Withdraw(userID string, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
	}
	job := w.TransactionJob{Type: w.JobWithdraw, Account: uid, Amount: amount, CallbackURL: callbackURL, JobID: uuid.New()}
	if t.Queue != nil {
		_ = t.Queue.QueueTransaction(job)
		return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"job_id": job.JobID.String(), "status": "pending"}}, nil
	}
	if err = t.DB.Transaction(uid, amount, "withdraw"); err != nil {
		if t.Events != nil {
			t.Events.PublishAsync(context.Background(), events.NewTransactionFailedEvent(uid, "withdraw", amount, err.Error(), "WITHDRAW_FAILED"))
		}
		return nil, err
	}
	if err = t.DB.Insert(&r.Transaction{AccountID: uid, Amount: amount, Type: "withdraw", Category: "debit", Description: "User withdraw"}); err != nil {
		return nil, err
	}
	if t.Events != nil {
		t.Events.PublishAsync(context.Background(), events.NewWithdrawCompletedEvent(uid, amount, fmt.Sprintf("wd_%s", uuid.New().String()[:8])))
	}
	return &ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"message": "Withdraw successfully"}}, nil
}

func (t *TransactionService) Transfer(userID string, amount decimal.Decimal, userTo, callbackURL string) (*ServiceResponse, error) {
	userFrom, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
	}
	job := w.TransactionJob{Type: w.JobTransfer, Account: userFrom, Amount: amount, ToEmail: userTo, CallbackURL: callbackURL, JobID: uuid.New()}
	if t.Queue != nil {
		_ = t.Queue.QueueTransaction(job)
		return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"job_id": job.JobID.String(), "status": "pending"}}, nil
	}
	foundUser, err := t.DB.FindUserByField("email", userTo)
	if err != nil {
		return nil, err
	}
	destinyUserID := foundUser.ID
	foundUserFrom, err := t.DB.FindUserByField("id", userFrom.String())
	if err != nil {
		return nil, err
	}
	if err := t.DB.Transaction(userFrom, amount, "withdraw"); err != nil {
		return nil, err
	}
	if err := t.DB.Transaction(destinyUserID, amount, "deposit"); err != nil {
		return nil, err
	}
	if err := t.DB.Insert(&r.Transaction{AccountID: userFrom, Amount: amount, Type: "transfer", Category: "debit", Description: "User transfer to " + userTo}); err != nil {
		return nil, err
	}
	if err := t.DB.Insert(&r.Transaction{AccountID: destinyUserID, Amount: amount, Type: "transfer", Category: "credit", Description: "User transfer from " + foundUserFrom.Email}); err != nil {
		return nil, err
	}
	if t.Events != nil {
		t.Events.PublishAsync(context.Background(), events.NewTransferCompletedEvent(userFrom, destinyUserID, amount, fmt.Sprintf("tr_%s", uuid.New().String()[:8])))
	}
	return &ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"message": "Transfer successfully"}}, nil
}

func (t *TransactionService) WithdrawTron(userID string, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
	}
	walletInfo, err := t.DB.GetWalletInfo(uid)
	if err != nil {
		return nil, fmt.Errorf("TRON wallet not found. Please contact support to generate your wallet")
	}
	destinationAddress := walletInfo.TronAddress
	// Uso do TronPort se disponível, senão fallback
	tronPort := t.Tron
	if tronPort == nil {
		return nil, fmt.Errorf("TRON service not configured")
	}
	if !tronPort.HasVaultConfigured() {
		return nil, fmt.Errorf("TRON vault is not configured. Please contact administrator")
	}
	vaultAddress := tronPort.GetVaultAddress()
	vaultPrivateKey := tronPort.GetVaultPrivateKey()
	amountInSun := amount.Mul(decimal.NewFromInt(1000000)).BigInt().Int64()
	txRecord := &r.Transaction{AccountID: uid, Amount: amount, Type: "withdraw", Category: "debit", Description: fmt.Sprintf("TRON withdraw to %s", destinationAddress), TronTxStatus: stringPtr("pending")}
	if err = t.DB.Insert(txRecord); err != nil {
		return nil, err
	}
	if callbackURL != "" {
		t.sendWebhook(callbackURL, map[string]interface{}{"status": "pending", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "description": "Withdraw request received and queued"})
	}
	if err = t.DB.Transaction(uid, amount, "withdraw"); err != nil {
		_ = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": "failed"})
		return nil, err
	}
	_ = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": "broadcasting"})
	if callbackURL != "" {
		t.sendWebhook(callbackURL, map[string]interface{}{"status": "broadcasting", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "description": "Broadcasting transaction to TRON network"})
	}
	txHash, sendError := tronPort.SendTransaction(vaultAddress, destinationAddress, amountInSun, vaultPrivateKey)
	if sendError != nil {
		_ = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": "failed"})
		if callbackURL != "" {
			t.sendWebhook(callbackURL, map[string]interface{}{"status": "failed", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "error": sendError.Error(), "description": "Transaction broadcast failed"})
		}
		return nil, fmt.Errorf("failed to broadcast transaction: %w", sendError)
	}
	_ = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_hash": txHash, "tron_tx_status": "broadcast_success"})
	if callbackURL != "" {
		t.sendWebhook(callbackURL, map[string]interface{}{"status": "broadcast_success", "tx_id": txRecord.ID.String(), "tx_hash": txHash, "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "from_address": vaultAddress, "timestamp": time.Now().Unix(), "description": "Transaction successfully broadcast to TRON network", "explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", txHash)})
	}
	if t.TronConfirm != nil {
		t.TronConfirm.SubmitConfirmationJob(uid, txRecord.ID, txHash, callbackURL)
	}
	return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"message": "Withdrawal broadcast to TRON blockchain", "tx_id": txRecord.ID.String(), "tx_hash": txHash, "amount": amount.String(), "to_address": destinationAddress, "status": "broadcast_success", "description": "Transaction is awaiting confirmations on TRON network", "explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", txHash)}}, nil
}

func stringPtr(s string) *string { return &s }

func (t *TransactionService) sendWebhook(url string, data map[string]interface{}) {
	if url == "" {
		return
	}
	data["event_type"], data["webhook_version"] = "tron_transaction_update", "1.0"
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "FinancialSystemPro/1.0")
		req.Header.Set("X-Webhook-Event", "tron_transaction_update")
		req.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", attempt))
		if secret := os.Getenv("WEBHOOK_SECRET"); secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(payload)
			req.Header.Set("X-Webhook-Signature", hex.EncodeToString(mac.Sum(nil)))
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
}

func (t *TransactionService) GetWalletInfo(userID uuid.UUID) (*r.WalletInfo, error) {
	return t.DB.GetWalletInfo(userID)
}
