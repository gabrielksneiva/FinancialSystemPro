package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Queue manager para Asynq
type QueueManager struct {
	client   *asynq.Client
	server   *asynq.Server
	logger   *zap.Logger
	redisURL string
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
func NewQueueManager(redisURL string, logger *zap.Logger) *QueueManager {
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

	// Cria client mas NÃO aguarda conexão (async)
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: opt.Addr})

	qm := &QueueManager{
		client:   client,
		logger:   logger,
		redisURL: redisURL,
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

		qm.logger.Warn(fmt.Sprintf("[REDIS DEBUG] attempt %d/%d failed to connect to redis", attempt, maxRetries),
			zap.Error(nil))

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
	var payload TransactionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		qm.logger.Error("failed to unmarshal deposit payload", zap.Error(err))
		return err
	}

	qm.logger.Info("processing deposit",
		zap.String("user_id", payload.UserID),
		zap.String("amount", payload.Amount),
	)

	// TODO: Implementar lógica real de depósito
	// Por enquanto, só logging
	return nil
}

func (qm *QueueManager) handleWithdraw(ctx context.Context, t *asynq.Task) error {
	var payload TransactionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		qm.logger.Error("failed to unmarshal withdraw payload", zap.Error(err))
		return err
	}

	qm.logger.Info("processing withdraw",
		zap.String("user_id", payload.UserID),
		zap.String("amount", payload.Amount),
	)

	// TODO: Implementar lógica real de saque
	return nil
}

func (qm *QueueManager) handleTransfer(ctx context.Context, t *asynq.Task) error {
	var payload TransactionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		qm.logger.Error("failed to unmarshal transfer payload", zap.Error(err))
		return err
	}

	qm.logger.Info("processing transfer",
		zap.String("from_user", payload.UserID),
		zap.String("to_email", payload.ToEmail),
		zap.String("amount", payload.Amount),
	)

	// TODO: Implementar lógica real de transferência
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
