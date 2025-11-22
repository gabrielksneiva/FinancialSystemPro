package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"financial-system-pro/repositories"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Queue manager para Asynq
type QueueManager struct {
	client   *asynq.Client
	server   *asynq.Server
	logger   *zap.Logger
	redisURL string
	database *repositories.NewDatabase
}

// Tipos de tarefas
const (
	TypeDeposit  = "transaction:deposit"
	TypeWithdraw = "transaction:withdraw"
	TypeTransfer = "transaction:transfer"
)

// Payload base para todas as transações
type TransactionPayload struct {
	UserID      string `json:"user_id"`
	Amount      string `json:"amount"`
	ToUserID    string `json:"to_user_id,omitempty"`
	ToEmail     string `json:"to_email,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// NewQueueManager cria um novo gerenciador de fila com retry para serverless
// Inicia de forma NÃO-BLOQUEANTE (async)
func NewQueueManager(redisURL string, logger *zap.Logger, database *repositories.NewDatabase) *QueueManager {
	// Log the configuration
	logger.Info("[REDIS DEBUG] initializing redis queue manager (async)",
		zap.String("redis_url_length", fmt.Sprintf("%d chars", len(redisURL))),
		zap.String("redis_url", redisURL))

	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("[REDIS DEBUG] failed to parse redis url",
			zap.Error(err),
			zap.String("redis_url", redisURL))
		return nil
	}

	logger.Info("[REDIS DEBUG] redis url parsed successfully",
		zap.String("addr", opt.Addr),
		zap.String("username", opt.Username),
		zap.Bool("has_password", opt.Password != ""))

	// Criar RedisClientOpt com todas as credenciais
	redisOpt := asynq.RedisClientOpt{
		Addr:     opt.Addr,
		Username: opt.Username,
		Password: opt.Password,
		DB:       opt.DB,
	}

	// Cria client mas NÃO aguarda conexão (async)
	client := asynq.NewClient(redisOpt)

	qm := &QueueManager{
		client:   client,
		logger:   logger,
		redisURL: redisURL,
		database: database,
	}

	// Tenta conectar em background com retry agressivo
	go qm.connectWithRetry()

	logger.Info("[REDIS DEBUG] queue manager created (will connect asynchronously)")
	return qm
}

// connectWithRetry tenta conectar ao Redis em background
func (qm *QueueManager) connectWithRetry() {
	maxRetries := 20
	var retryDelay time.Duration = 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := qm.client.Ping(); err == nil {
			qm.logger.Info("[REDIS DEBUG] successfully connected to redis queue (async)",
				zap.Int("attempt", attempt))
			return
		}

		qm.logger.Warn(fmt.Sprintf("[REDIS DEBUG] attempt %d/%d failed to connect to redis", attempt, maxRetries))

		if attempt < maxRetries {
			qm.logger.Debug(fmt.Sprintf("[REDIS DEBUG] retrying in %v...", retryDelay))
			time.Sleep(retryDelay)
			retryDelay *= 2 // exponential backoff
		}
	}

	qm.logger.Warn("[REDIS DEBUG] failed to connect to redis after all retries, running without async queue")
}

// StartWorkers inicia os workers para processar tarefas
func (qm *QueueManager) StartWorkers(ctx context.Context) error {
	opt, err := redis.ParseURL(qm.redisURL)
	if err != nil {
		return err
	}

	// Criar RedisClientOpt com todas as credenciais
	redisOpt := asynq.RedisClientOpt{
		Addr:     opt.Addr,
		Username: opt.Username,
		Password: opt.Password,
		DB:       opt.DB,
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default":      10,
				"critical":     5,
				"transactions": 10,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				if n == 0 {
					return 0
				}
				return time.Duration(n*n) * time.Minute
			},
		},
	)

	qm.server = srv

	// Registrar handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeDeposit, qm.handleDeposit)
	mux.HandleFunc(TypeWithdraw, qm.handleWithdraw)
	mux.HandleFunc(TypeTransfer, qm.handleTransfer)

	qm.logger.Info("starting async workers", zap.Int("concurrency", 10))

	go func() {
		if err := srv.Run(mux); err != nil {
			qm.logger.Error("failed to run server", zap.Error(err))
		}
	}()

	return nil
}

// EnqueueDeposit enfileira uma tarefa de depósito
func (qm *QueueManager) EnqueueDeposit(ctx context.Context, userID, amount, callbackURL string) (string, error) {
	payload := TransactionPayload{
		UserID:      userID,
		Amount:      amount,
		CallbackURL: callbackURL,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		qm.logger.Error("failed to marshal deposit payload", zap.Error(err))
		return "", err
	}

	task := asynq.NewTask(TypeDeposit, data)
	info, err := qm.client.EnqueueContext(ctx, task, asynq.Queue("transactions"), asynq.MaxRetry(3))
	if err != nil {
		qm.logger.Error("failed to enqueue deposit", zap.Error(err))
		return "", err
	}

	qm.logger.Info("deposit task enqueued",
		zap.String("task_id", info.ID),
		zap.String("user_id", userID),
		zap.String("amount", amount),
	)
	return info.ID, nil
}

// EnqueueWithdraw enfileira uma tarefa de saque
func (qm *QueueManager) EnqueueWithdraw(ctx context.Context, userID, amount, callbackURL string) (string, error) {
	payload := TransactionPayload{
		UserID:      userID,
		Amount:      amount,
		CallbackURL: callbackURL,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		qm.logger.Error("failed to marshal withdraw payload", zap.Error(err))
		return "", err
	}

	task := asynq.NewTask(TypeWithdraw, data)
	info, err := qm.client.EnqueueContext(ctx, task, asynq.Queue("transactions"), asynq.MaxRetry(3))
	if err != nil {
		qm.logger.Error("failed to enqueue withdraw", zap.Error(err))
		return "", err
	}

	qm.logger.Info("withdraw task enqueued",
		zap.String("task_id", info.ID),
		zap.String("user_id", userID),
	)
	return info.ID, nil
}

// EnqueueTransfer enfileira uma tarefa de transferência
func (qm *QueueManager) EnqueueTransfer(ctx context.Context, userID, amount, toEmail, callbackURL string) (string, error) {
	payload := TransactionPayload{
		UserID:      userID,
		Amount:      amount,
		ToEmail:     toEmail,
		CallbackURL: callbackURL,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		qm.logger.Error("failed to marshal transfer payload", zap.Error(err))
		return "", err
	}

	task := asynq.NewTask(TypeTransfer, data)
	info, err := qm.client.EnqueueContext(ctx, task, asynq.Queue("transactions"), asynq.MaxRetry(3))
	if err != nil {
		qm.logger.Error("failed to enqueue transfer", zap.Error(err))
		return "", err
	}

	qm.logger.Info("transfer task enqueued",
		zap.String("task_id", info.ID),
		zap.String("user_id", userID),
		zap.String("to_email", toEmail),
	)
	return info.ID, nil
}

// Handlers para processar tarefas
func (qm *QueueManager) handleDeposit(ctx context.Context, t *asynq.Task) error {
	if qm.database == nil {
		qm.logger.Warn("database not available for deposit processing")
		return fmt.Errorf("database not available")
	}

	var payload TransactionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		qm.logger.Error("failed to unmarshal deposit payload", zap.Error(err))
		return err
	}

	// Parse userID
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		qm.logger.Error("invalid user id format", zap.String("user_id", payload.UserID), zap.Error(err))
		return fmt.Errorf("invalid user id format: %w", err)
	}

	// Validar amount
	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil {
		qm.logger.Error("invalid amount format", zap.String("amount", payload.Amount), zap.Error(err))
		return fmt.Errorf("invalid amount format: %w", err)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		qm.logger.Warn("invalid deposit amount", zap.String("user_id", payload.UserID), zap.Stringer("amount", amount))
		return fmt.Errorf("deposit amount must be greater than zero")
	}

	// Executar depósito
	qm.logger.Info("processing deposit",
		zap.String("user_id", payload.UserID),
		zap.String("amount", payload.Amount),
	)

	if err := qm.database.Transaction(userID, amount, "deposit"); err != nil {
		qm.logger.Error("deposit transaction failed",
			zap.String("user_id", payload.UserID),
			zap.String("amount", payload.Amount),
			zap.Error(err))
		return fmt.Errorf("deposit transaction failed: %w", err)
	}

	// Registrar na tabela transactions
	if err := qm.database.Insert(&repositories.Transaction{
		AccountID:   userID,
		Amount:      amount,
		Type:        "deposit",
		Category:    "credit",
		Description: "Deposit to account",
	}); err != nil {
		qm.logger.Error("failed to insert deposit record",
			zap.String("user_id", payload.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to record deposit: %w", err)
	}

	qm.logger.Info("deposit completed successfully",
		zap.String("user_id", payload.UserID),
		zap.String("amount", payload.Amount),
	)

	return nil
}

func (qm *QueueManager) handleWithdraw(ctx context.Context, t *asynq.Task) error {
	if qm.database == nil {
		qm.logger.Warn("database not available for withdraw processing")
		return fmt.Errorf("database not available")
	}

	var payload TransactionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		qm.logger.Error("failed to unmarshal withdraw payload", zap.Error(err))
		return err
	}

	// Parse userID
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		qm.logger.Error("invalid user id format", zap.String("user_id", payload.UserID), zap.Error(err))
		return fmt.Errorf("invalid user id format: %w", err)
	}

	// Validar amount
	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil {
		qm.logger.Error("invalid amount format", zap.String("amount", payload.Amount), zap.Error(err))
		return fmt.Errorf("invalid amount format: %w", err)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		qm.logger.Warn("invalid withdraw amount", zap.String("user_id", payload.UserID), zap.Stringer("amount", amount))
		return fmt.Errorf("withdraw amount must be greater than zero")
	}

	// Verificar saldo
	balance, err := qm.database.Balance(userID)
	if err != nil {
		qm.logger.Error("failed to check balance",
			zap.String("user_id", payload.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to check balance: %w", err)
	}

	if balance.LessThan(amount) {
		qm.logger.Warn("insufficient balance",
			zap.String("user_id", payload.UserID),
			zap.Stringer("balance", balance),
			zap.Stringer("requested", amount))
		return fmt.Errorf("insufficient balance: have %s, need %s", balance, amount)
	}

	// Executar saque
	qm.logger.Info("processing withdraw",
		zap.String("user_id", payload.UserID),
		zap.String("amount", payload.Amount),
	)

	if err := qm.database.Transaction(userID, amount, "withdraw"); err != nil {
		qm.logger.Error("withdraw transaction failed",
			zap.String("user_id", payload.UserID),
			zap.String("amount", payload.Amount),
			zap.Error(err))
		return fmt.Errorf("withdraw transaction failed: %w", err)
	}

	// Registrar na tabela transactions
	if err := qm.database.Insert(&repositories.Transaction{
		AccountID:   userID,
		Amount:      amount,
		Type:        "withdraw",
		Category:    "debit",
		Description: "Withdrawal from account",
	}); err != nil {
		qm.logger.Error("failed to insert withdraw record",
			zap.String("user_id", payload.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to record withdrawal: %w", err)
	}

	qm.logger.Info("withdraw completed successfully",
		zap.String("user_id", payload.UserID),
		zap.String("amount", payload.Amount),
	)

	return nil
}

func (qm *QueueManager) handleTransfer(ctx context.Context, t *asynq.Task) error {
	if qm.database == nil {
		qm.logger.Warn("database not available for transfer processing")
		return fmt.Errorf("database not available")
	}

	var payload TransactionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		qm.logger.Error("failed to unmarshal transfer payload", zap.Error(err))
		return err
	}

	// Parse userID
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		qm.logger.Error("invalid user id format", zap.String("user_id", payload.UserID), zap.Error(err))
		return fmt.Errorf("invalid user id format: %w", err)
	}

	// Validar amount
	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil {
		qm.logger.Error("invalid amount format", zap.String("amount", payload.Amount), zap.Error(err))
		return fmt.Errorf("invalid amount format: %w", err)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		qm.logger.Warn("invalid transfer amount", zap.String("user_id", payload.UserID), zap.Stringer("amount", amount))
		return fmt.Errorf("transfer amount must be greater than zero")
	}

	// Verificar saldo do remetente
	balance, err := qm.database.Balance(userID)
	if err != nil {
		qm.logger.Error("failed to check sender balance",
			zap.String("user_id", payload.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to check sender balance: %w", err)
	}

	if balance.LessThan(amount) {
		qm.logger.Warn("insufficient balance for transfer",
			zap.String("user_id", payload.UserID),
			zap.Stringer("balance", balance),
			zap.Stringer("requested", amount))
		return fmt.Errorf("insufficient balance: have %s, need %s", balance, amount)
	}

	// Encontrar usuário destinatário
	recipient, err := qm.database.FindUserByField("email", payload.ToEmail)
	if err != nil {
		qm.logger.Warn("recipient not found",
			zap.String("to_email", payload.ToEmail),
			zap.Error(err))
		return fmt.Errorf("recipient not found: %w", err)
	}

	qm.logger.Info("processing transfer",
		zap.String("from_user", payload.UserID),
		zap.String("to_email", payload.ToEmail),
		zap.String("amount", payload.Amount),
	)

	// Débito do remetente
	if err := qm.database.Transaction(userID, amount, "withdraw"); err != nil {
		qm.logger.Error("transfer withdraw failed",
			zap.String("from_id", payload.UserID),
			zap.Error(err))
		return fmt.Errorf("transfer withdraw failed: %w", err)
	}

	// Crédito do destinatário
	if err := qm.database.Transaction(recipient.ID, amount, "deposit"); err != nil {
		qm.logger.Error("transfer deposit failed",
			zap.String("to_id", recipient.ID.String()),
			zap.Error(err))
		return fmt.Errorf("transfer deposit failed: %w", err)
	}

	// Registrar transferência (débito)
	if err := qm.database.Insert(&repositories.Transaction{
		AccountID:   userID,
		Amount:      amount,
		Type:        "transfer",
		Category:    "debit",
		Description: "Transfer to " + payload.ToEmail,
	}); err != nil {
		qm.logger.Error("failed to insert transfer debit record",
			zap.String("user_id", payload.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to record transfer: %w", err)
	}

	// Registrar transferência (crédito)
	if err := qm.database.Insert(&repositories.Transaction{
		AccountID:   recipient.ID,
		Amount:      amount,
		Type:        "transfer",
		Category:    "credit",
		Description: "Transfer from " + payload.UserID,
	}); err != nil {
		qm.logger.Error("failed to insert transfer credit record",
			zap.String("user_id", recipient.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to record transfer receipt: %w", err)
	}

	qm.logger.Info("transfer completed successfully",
		zap.String("from_user", payload.UserID),
		zap.String("to_user", recipient.ID.String()),
		zap.String("amount", payload.Amount),
	)

	return nil
}

// Close fecha a conexão com Redis
func (qm *QueueManager) Close() error {
	if qm.server != nil {
		qm.server.Stop()
	}
	return qm.client.Close()
}

// GetTaskInfo retorna informações sobre uma tarefa
func (qm *QueueManager) GetTaskInfo(ctx context.Context, queue, taskID string) (*asynq.TaskInfo, error) {
	opt, err := redis.ParseURL(qm.redisURL)
	if err != nil {
		return nil, err
	}

	// Criar RedisClientOpt com todas as credenciais
	redisOpt := asynq.RedisClientOpt{
		Addr:     opt.Addr,
		Username: opt.Username,
		Password: opt.Password,
		DB:       opt.DB,
	}

	inspector := asynq.NewInspector(redisOpt)
	defer inspector.Close()

	info, err := inspector.GetTaskInfo(queue, taskID)
	if err != nil {
		return nil, err
	}

	return info, nil
}
