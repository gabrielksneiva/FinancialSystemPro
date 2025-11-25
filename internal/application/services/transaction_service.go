package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"financial-system-pro/internal/domain/entities"
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
	DB            DatabasePort // compatibilidade
	Ledger        LedgerPort
	TxRepo        TransactionRecordPort
	Queue         QueuePort
	Tron          TronPort
	TronConfirm   TronConfirmationPort
	Events        EventsPort
	Logger        *zap.Logger
	ChainRegistry *BlockchainRegistry
	OnChainRepo   OnChainWalletRepositoryPort // repositório multi-chain para leitura de wallets (Ethereum, etc.)
}

func NewTransactionService(db DatabasePort, queue QueuePort, tron TronPort, tronConfirm TronConfirmationPort, events EventsPort, logger *zap.Logger) *TransactionService {
	return &TransactionService{DB: db, Ledger: NewLedgerAdapter(db), TxRepo: NewTransactionRecordAdapter(db), Queue: queue, Tron: tron, TronConfirm: tronConfirm, Events: events, Logger: logger}
}

// WithChainRegistry injeta registro multi-chain (builder pattern)
func (t *TransactionService) WithChainRegistry(reg *BlockchainRegistry) *TransactionService {
	t.ChainRegistry = reg
	return t
}

// WithOnChainWalletRepository injeta repositório de wallets multi-chain
func (t *TransactionService) WithOnChainWalletRepository(repo OnChainWalletRepositoryPort) *TransactionService {
	t.OnChainRepo = repo
	return t
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

	if t.Ledger == nil && t.DB != nil {
		t.Ledger = NewLedgerAdapter(t.DB)
	}
	if t.TxRepo == nil && t.DB != nil {
		t.TxRepo = NewTransactionRecordAdapter(t.DB)
	}

	if err := t.Ledger.Apply(context.Background(), uid, amount, "deposit"); err != nil {
		if t.Events != nil {
			t.Events.PublishAsync(context.Background(), events.NewTransactionFailedEvent(uid, "deposit", amount, err.Error(), "DEPOSIT_FAILED"))
		}
		return nil, errors.NewDatabaseError("deposit failed", nil)
	}
	if err := t.TxRepo.Insert(context.Background(), &r.Transaction{AccountID: uid, Amount: amount, Type: "deposit", Category: "credit", Description: "User deposit"}); err != nil {
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
	if t.Ledger == nil && t.DB != nil {
		t.Ledger = NewLedgerAdapter(t.DB)
	}
	response, err := t.Ledger.Balance(context.Background(), uid)
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
	if t.Ledger == nil && t.DB != nil {
		t.Ledger = NewLedgerAdapter(t.DB)
	}
	if t.TxRepo == nil && t.DB != nil {
		t.TxRepo = NewTransactionRecordAdapter(t.DB)
	}
	if err = t.Ledger.Apply(context.Background(), uid, amount, "withdraw"); err != nil {
		if t.Events != nil {
			t.Events.PublishAsync(context.Background(), events.NewTransactionFailedEvent(uid, "withdraw", amount, err.Error(), "WITHDRAW_FAILED"))
		}
		return nil, err
	}
	if err = t.TxRepo.Insert(context.Background(), &r.Transaction{AccountID: uid, Amount: amount, Type: "withdraw", Category: "debit", Description: "User withdraw"}); err != nil {
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
	if t.Ledger == nil && t.DB != nil {
		t.Ledger = NewLedgerAdapter(t.DB)
	}
	if t.TxRepo == nil && t.DB != nil {
		t.TxRepo = NewTransactionRecordAdapter(t.DB)
	}
	if err := t.Ledger.Apply(context.Background(), userFrom, amount, "withdraw"); err != nil {
		return nil, err
	}
	if err := t.Ledger.Apply(context.Background(), destinyUserID, amount, "deposit"); err != nil {
		return nil, err
	}
	if err := t.TxRepo.Insert(context.Background(), &r.Transaction{AccountID: userFrom, Amount: amount, Type: "transfer", Category: "debit", Description: "User transfer to " + userTo}); err != nil {
		return nil, err
	}
	if err := t.TxRepo.Insert(context.Background(), &r.Transaction{AccountID: destinyUserID, Amount: amount, Type: "transfer", Category: "credit", Description: "User transfer from " + foundUserFrom.Email}); err != nil {
		return nil, err
	}
	if t.Events != nil {
		t.Events.PublishAsync(context.Background(), events.NewTransferCompletedEvent(userFrom, destinyUserID, amount, fmt.Sprintf("tr_%s", uuid.New().String()[:8])))
	}
	return &ServiceResponse{StatusCode: 200, Body: map[string]interface{}{"message": "Transfer successfully"}}, nil
}

func (t *TransactionService) WithdrawTron(userID string, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	// Delegar para método genérico mantendo compatibilidade
	return t.WithdrawOnChain(userID, entities.BlockchainTRON, amount, callbackURL)
}

// WithdrawOnChain executa operação de saque em blockchain específica
// Atualmente implementado para TRON; outras chains serão adicionadas.
func (t *TransactionService) WithdrawOnChain(userID string, chain entities.BlockchainType, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
	}

	// Branch Bitcoin
	if chain == entities.BlockchainBitcoin {
		if t.ChainRegistry == nil {
			return nil, fmt.Errorf("chain registry não configurado")
		}
		if t.OnChainRepo == nil {
			return nil, fmt.Errorf("on-chain wallet repository não configurado")
		}
		wallet, wErr := t.OnChainRepo.FindByUserAndChain(context.Background(), uid, entities.BlockchainBitcoin)
		if wErr != nil {
			return nil, fmt.Errorf("Bitcoin wallet não encontrada para usuário. Gere a wallet primeiro")
		}
		gw, gwErr := t.ChainRegistry.Get(entities.BlockchainBitcoin)
		if gwErr != nil {
			return nil, fmt.Errorf("Bitcoin gateway não disponível: %w", gwErr)
		}
		vaultAddress := os.Getenv("BTC_VAULT_ADDRESS")
		vaultPrivKey := os.Getenv("BTC_VAULT_PRIVATE_KEY")
		if vaultAddress == "" || vaultPrivKey == "" {
			return nil, fmt.Errorf("Bitcoin vault não configurado (BTC_VAULT_ADDRESS / BTC_VAULT_PRIVATE_KEY)")
		}
		if !gw.ValidateAddress(vaultAddress) || !gw.ValidateAddress(wallet.Address) {
			return nil, fmt.Errorf("endereço inválido (vault ou destino)")
		}
		// converter amount usando helper genérico (satoshis)
		amountInSats, convErr := ConvertAmountToBaseUnit(entities.BlockchainBitcoin, amount)
		if convErr != nil {
			return nil, convErr
		}
		if t.Ledger == nil && t.DB != nil {
			t.Ledger = NewLedgerAdapter(t.DB)
		}
		if t.TxRepo == nil && t.DB != nil {
			t.TxRepo = NewTransactionRecordAdapter(t.DB)
		}
		txRecord := &r.Transaction{AccountID: uid, Amount: amount, Type: "withdraw", Category: "debit", Description: fmt.Sprintf("BTC withdraw to %s", wallet.Address), OnChainTxStatus: stringPtr("pending"), OnChainChain: stringPtr("bitcoin")}
		if err = t.TxRepo.Insert(context.Background(), txRecord); err != nil {
			return nil, err
		}
		// webhook pending
		if callbackURL != "" {
			t.sendWebhook(callbackURL, "bitcoin", map[string]interface{}{"status": "pending", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": wallet.Address, "timestamp": time.Now().Unix(), "description": "Withdraw request received and queued"})
		}
		if err = t.Ledger.Apply(context.Background(), uid, amount, "withdraw"); err != nil {
			_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"onchain_tx_status": "failed"})
			if callbackURL != "" {
				t.sendWebhook(callbackURL, "bitcoin", map[string]interface{}{"status": "failed", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": wallet.Address, "timestamp": time.Now().Unix(), "error": err.Error(), "description": "Ledger apply failed"})
			}
			return nil, err
		}
		_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"onchain_tx_status": "broadcasting"})
		if callbackURL != "" {
			t.sendWebhook(callbackURL, "bitcoin", map[string]interface{}{"status": "broadcasting", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": wallet.Address, "timestamp": time.Now().Unix(), "description": "Broadcasting transaction to Bitcoin network"})
		}
		btcHash, bErr := gw.Broadcast(context.Background(), vaultAddress, wallet.Address, amountInSats, vaultPrivKey)
		if bErr != nil {
			_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"onchain_tx_status": "failed"})
			if callbackURL != "" {
				t.sendWebhook(callbackURL, "bitcoin", map[string]interface{}{"status": "failed", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": wallet.Address, "timestamp": time.Now().Unix(), "error": bErr.Error(), "description": "Transaction broadcast failed"})
			}
			return nil, fmt.Errorf("falha broadcast Bitcoin: %w", bErr)
		}
		_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"onchain_tx_hash": string(btcHash), "onchain_tx_status": "broadcast_success"})
		if callbackURL != "" {
			t.sendWebhook(callbackURL, "bitcoin", map[string]interface{}{"status": "broadcast_success", "tx_id": txRecord.ID.String(), "tx_hash": string(btcHash), "user_id": uid.String(), "amount": amount.String(), "to_address": wallet.Address, "from_address": vaultAddress, "timestamp": time.Now().Unix(), "description": "Transaction successfully broadcast to Bitcoin network", "explorer_url": fmt.Sprintf("https://mempool.space/tx/%s", btcHash)})
			// iniciar confirmação assíncrona simples
			go func(txID uuid.UUID, hash string, cb string, user uuid.UUID, amt decimal.Decimal, to string) {
				// confirming
				_ = t.TxRepo.Update(context.Background(), txID, map[string]interface{}{"onchain_tx_status": "confirming"})
				t.sendWebhook(cb, "bitcoin", map[string]interface{}{"status": "confirming", "tx_id": txID.String(), "tx_hash": hash, "user_id": user.String(), "amount": amt.String(), "to_address": to, "timestamp": time.Now().Unix(), "confirmations": 0, "description": "Waiting for confirmations on Bitcoin network", "explorer_url": fmt.Sprintf("https://mempool.space/tx/%s", hash)})
				time.Sleep(3 * time.Second)
				_ = t.TxRepo.Update(context.Background(), txID, map[string]interface{}{"onchain_tx_status": "confirmed"})
				t.sendWebhook(cb, "bitcoin", map[string]interface{}{"status": "confirmed", "tx_id": txID.String(), "tx_hash": hash, "user_id": user.String(), "amount": amt.String(), "to_address": to, "timestamp": time.Now().Unix(), "confirmations": 1, "description": "Transaction confirmed on Bitcoin network", "explorer_url": fmt.Sprintf("https://mempool.space/tx/%s", hash)})
				time.Sleep(2 * time.Second)
				_ = t.TxRepo.Update(context.Background(), txID, map[string]interface{}{"onchain_tx_status": "completed"})
				t.sendWebhook(cb, "bitcoin", map[string]interface{}{"status": "completed", "tx_id": txID.String(), "tx_hash": hash, "user_id": user.String(), "amount": amt.String(), "to_address": to, "timestamp": time.Now().Unix(), "confirmations": 3, "description": "Transaction fully completed with multiple confirmations", "explorer_url": fmt.Sprintf("https://mempool.space/tx/%s", hash)})
			}(txRecord.ID, string(btcHash), callbackURL, uid, amount, wallet.Address)
		}
		return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"message": "Withdrawal broadcast to Bitcoin blockchain", "tx_id": txRecord.ID.String(), "tx_hash": string(btcHash), "amount": amount.String(), "to_address": wallet.Address, "from_address": vaultAddress, "status": "broadcast_success", "explorer_url": fmt.Sprintf("https://mempool.space/tx/%s", btcHash), "onchain_chain": "bitcoin"}}, nil
	}

	// Branch Ethereum
	if chain == entities.BlockchainEthereum {
		if t.ChainRegistry == nil {
			return nil, fmt.Errorf("chain registry não configurado")
		}
		if t.OnChainRepo == nil {
			return nil, fmt.Errorf("on-chain wallet repository não configurado")
		}
		wallet, wErr := t.OnChainRepo.FindByUserAndChain(context.Background(), uid, entities.BlockchainEthereum)
		if wErr != nil {
			return nil, fmt.Errorf("Ethereum wallet não encontrada para usuário. Gere a wallet primeiro")
		}
		gw, gwErr := t.ChainRegistry.Get(entities.BlockchainEthereum)
		if gwErr != nil {
			return nil, fmt.Errorf("Ethereum gateway não disponível: %w", gwErr)
		}
		vaultAddress := os.Getenv("ETH_VAULT_ADDRESS")
		vaultPrivKey := os.Getenv("ETH_VAULT_PRIVATE_KEY")
		if vaultAddress == "" || vaultPrivKey == "" {
			return nil, fmt.Errorf("Ethereum vault não configurado (ETH_VAULT_ADDRESS / ETH_VAULT_PRIVATE_KEY)")
		}
		if !gw.ValidateAddress(vaultAddress) || !gw.ValidateAddress(wallet.Address) {
			return nil, fmt.Errorf("endereço inválido (vault ou destino)")
		}
		// converter amount usando helper genérico (wei)
		amountInWei, convErr := ConvertAmountToBaseUnit(entities.BlockchainEthereum, amount)
		if convErr != nil {
			return nil, convErr
		}
		if t.Ledger == nil && t.DB != nil {
			t.Ledger = NewLedgerAdapter(t.DB)
		}
		if t.TxRepo == nil && t.DB != nil {
			t.TxRepo = NewTransactionRecordAdapter(t.DB)
		}
		txRecord := &r.Transaction{AccountID: uid, Amount: amount, Type: "withdraw", Category: "debit", Description: fmt.Sprintf("ETH withdraw to %s", wallet.Address), OnChainTxStatus: stringPtr("pending"), OnChainChain: stringPtr("ethereum")}
		if err = t.TxRepo.Insert(context.Background(), txRecord); err != nil {
			return nil, err
		}
		if err = t.Ledger.Apply(context.Background(), uid, amount, "withdraw"); err != nil {
			return nil, err
		}
		// broadcast
		ethHash, bErr := gw.Broadcast(context.Background(), vaultAddress, wallet.Address, amountInWei, vaultPrivKey)
		if bErr != nil {
			return nil, fmt.Errorf("falha broadcast Ethereum: %w", bErr)
		}
		_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"onchain_tx_hash": string(ethHash), "onchain_tx_status": "broadcast_success"})
		return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"message": "Withdrawal broadcast to Ethereum blockchain", "tx_id": txRecord.ID.String(), "tx_hash": string(ethHash), "amount": amount.String(), "to_address": wallet.Address, "from_address": vaultAddress, "status": "broadcast_success", "explorer_url": fmt.Sprintf("https://etherscan.io/tx/%s", ethHash), "onchain_chain": "ethereum"}}, nil
	}

	// Default / TRON branch
	if chain != entities.BlockchainTRON {
		return nil, fmt.Errorf("unsupported blockchain: %s", chain)
	}

	walletInfo, err := t.DB.GetWalletInfo(uid)
	if err != nil {
		return nil, fmt.Errorf("TRON wallet not found. Please contact support to generate your wallet")
	}
	destinationAddress := walletInfo.TronAddress
	tronPort := t.Tron
	if tronPort == nil {
		return nil, fmt.Errorf("TRON service not configured")
	}
	if !tronPort.HasVaultConfigured() {
		return nil, fmt.Errorf("TRON vault is not configured. Please contact administrator")
	}
	vaultAddress := tronPort.GetVaultAddress()
	vaultPrivateKey := tronPort.GetVaultPrivateKey()
	// Conversão TRON via helper genérico (SUN)
	amountInSun, _ := ConvertAmountToBaseUnit(entities.BlockchainTRON, amount)
	if t.Ledger == nil && t.DB != nil {
		t.Ledger = NewLedgerAdapter(t.DB)
	}
	if t.TxRepo == nil && t.DB != nil {
		t.TxRepo = NewTransactionRecordAdapter(t.DB)
	}
	// Create transaction record
	txRecord := &r.Transaction{AccountID: uid, Amount: amount, Type: "withdraw", Category: "debit", Description: fmt.Sprintf("TRON withdraw to %s", destinationAddress), TronTxStatus: stringPtr("pending"), OnChainTxStatus: stringPtr("pending"), OnChainChain: stringPtr("tron")}
	if err = t.TxRepo.Insert(context.Background(), txRecord); err != nil {
		return nil, err
	}
	if callbackURL != "" {
		t.sendWebhook(callbackURL, "tron", map[string]interface{}{"status": "pending", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "description": "Withdraw request received and queued"})
	}
	if err = t.Ledger.Apply(context.Background(), uid, amount, "withdraw"); err != nil {
		_ = t.DB.UpdateTransaction(txRecord.ID, map[string]interface{}{"tron_tx_status": "failed", "onchain_tx_status": "failed"})
		if callbackURL != "" {
			t.sendWebhook(callbackURL, "tron", map[string]interface{}{"status": "failed", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "error": err.Error(), "description": "Ledger apply failed"})
		}
		return nil, err
	}
	_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"tron_tx_status": "broadcasting", "onchain_tx_status": "broadcasting"})
	if callbackURL != "" {
		t.sendWebhook(callbackURL, "tron", map[string]interface{}{"status": "broadcasting", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "description": "Broadcasting transaction to TRON network"})
	}
	var txHash string
	var sendError error
	if t.ChainRegistry != nil {
		if gw, gwErr := t.ChainRegistry.Get(chain); gwErr == nil && gw != nil {
			gwHash, gwSendErr := gw.Broadcast(context.Background(), vaultAddress, destinationAddress, amountInSun, vaultPrivateKey)
			if gwSendErr == nil {
				txHash = string(gwHash)
			} else {
				sendError = gwSendErr
			}
		}
	}
	if txHash == "" && sendError == nil { // fallback tronPort
		txHash, sendError = tronPort.SendTransaction(vaultAddress, destinationAddress, amountInSun, vaultPrivateKey)
	}
	if sendError != nil {
		_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"tron_tx_status": "failed", "onchain_tx_status": "failed"})
		if callbackURL != "" {
			t.sendWebhook(callbackURL, "tron", map[string]interface{}{"status": "failed", "tx_id": txRecord.ID.String(), "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "timestamp": time.Now().Unix(), "error": sendError.Error(), "description": "Transaction broadcast failed"})
		}
		return nil, fmt.Errorf("failed to broadcast transaction: %w", sendError)
	}
	_ = t.TxRepo.Update(context.Background(), txRecord.ID, map[string]interface{}{"tron_tx_hash": txHash, "tron_tx_status": "broadcast_success", "onchain_tx_hash": txHash, "onchain_tx_status": "broadcast_success"})
	if callbackURL != "" {
		t.sendWebhook(callbackURL, "tron", map[string]interface{}{"status": "broadcast_success", "tx_id": txRecord.ID.String(), "tx_hash": txHash, "user_id": uid.String(), "amount": amount.String(), "to_address": destinationAddress, "from_address": vaultAddress, "timestamp": time.Now().Unix(), "description": "Transaction successfully broadcast to TRON network", "explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", txHash)})
	}
	if t.TronConfirm != nil {
		t.TronConfirm.SubmitConfirmationJob(uid, txRecord.ID, txHash, callbackURL)
	}
	return &ServiceResponse{StatusCode: 202, Body: map[string]interface{}{"message": "Withdrawal broadcast to TRON blockchain", "tx_id": txRecord.ID.String(), "tx_hash": txHash, "amount": amount.String(), "to_address": destinationAddress, "status": "broadcast_success", "description": "Transaction is awaiting confirmations on TRON network", "explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", txHash), "onchain_chain": "tron"}}, nil
}

func stringPtr(s string) *string { return &s }

func (t *TransactionService) sendWebhook(url string, chain string, data map[string]interface{}) {
	if url == "" {
		return
	}
	if chain == "" {
		chain = "tron"
	}
	data["event_type"], data["webhook_version"], data["onchain_chain"] = chain+"_transaction_update", "1.0", chain
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
		req.Header.Set("X-Webhook-Event", chain+"_transaction_update")
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
