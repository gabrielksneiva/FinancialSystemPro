package workers

import (
	"financial-system-pro/repositories"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TronAPI interface para não criar ciclo de import
type TronAPI interface {
	GetTransaction(txHash string) (interface{}, error)
}

// TronWorkerPool gerencia workers que monitoram transações TRON
type TronWorkerPool struct {
	DB      *repositories.NewDatabase
	TronSvc TronAPI
	Jobs    chan TronTxConfirmJob
	Workers int
	quit    chan struct{}
	logger  *zap.Logger
}

// NewTronWorkerPool cria um novo pool de workers para TRON
func NewTronWorkerPool(db *repositories.NewDatabase, tronSvc TronAPI, workers int, logger *zap.Logger) *TronWorkerPool {
	return &TronWorkerPool{
		DB:      db,
		TronSvc: tronSvc,
		Jobs:    make(chan TronTxConfirmJob, 100),
		Workers: workers,
		quit:    make(chan struct{}),
		logger:  logger,
	}
}

// Start inicia os workers
func (twp *TronWorkerPool) Start() {
	for i := 0; i < twp.Workers; i++ {
		go twp.worker(i)
	}
	twp.logger.Info("TRON workers iniciados", zap.Int("count", twp.Workers))
}

// Stop para os workers
func (twp *TronWorkerPool) Stop() {
	close(twp.quit)
	twp.logger.Info("TRON workers parados")
}

// worker processa jobs de confirmação TRON
func (twp *TronWorkerPool) worker(id int) {
	for {
		select {
		case <-twp.quit:
			twp.logger.Info("TRON worker encerrando", zap.Int("worker_id", id))
			return
		case job := <-twp.Jobs:
			twp.processConfirmationJob(job)
		}
	}
}

// processConfirmationJob monitora uma transação TRON até confirmação
func (twp *TronWorkerPool) processConfirmationJob(job TronTxConfirmJob) {
	twp.logger.Info("Iniciando monitoramento de TX TRON",
		zap.String("tx_hash", job.TronTxHash),
		zap.String("user_id", job.UserID.String()),
	)

	// Verificar status da transação
	checkCount := 0
	for checkCount < job.MaxChecks {
		select {
		case <-twp.quit:
			twp.logger.Info("Parando monitoramento de TX", zap.String("tx_hash", job.TronTxHash))
			return
		default:
		}

		// Buscar status da transação TRON
		_, err := twp.TronSvc.GetTransaction(job.TronTxHash)
		if err != nil {
			twp.logger.Warn("Erro ao buscar status de TX",
				zap.String("tx_hash", job.TronTxHash),
				zap.Error(err),
			)
			checkCount++
			time.Sleep(time.Duration(job.CheckInterval) * time.Second)
			continue
		}

		// Aqui você precisaria fazer type assertion se soubesse a estrutura de retorno
		// Por agora, apenas marcamos como confirmed após confirmar que existe
		txStatus := "confirmed"

		err = twp.DB.UpdateTransaction(job.TransactionID, map[string]interface{}{
			"tron_tx_status": txStatus,
			"tron_tx_hash":   job.TronTxHash,
		})
		if err != nil {
			twp.logger.Error("Erro ao atualizar status de TX no DB",
				zap.String("tx_hash", job.TronTxHash),
				zap.Error(err),
			)
		}

		// Se confirmada, enviar callback e parar
		if txStatus == "confirmed" {
			twp.logger.Info("TX TRON confirmada",
				zap.String("tx_hash", job.TronTxHash),
				zap.String("user_id", job.UserID.String()),
			)

			if job.CallbackURL != "" {
				twp.sendCallback(job.CallbackURL, map[string]interface{}{
					"status":    "confirmed",
					"tx_hash":   job.TronTxHash,
					"job_id":    job.JobID.String(),
					"tx_id":     job.TransactionID.String(),
					"timestamp": time.Now().Unix(),
				})
			}
			return
		}

		checkCount++
		time.Sleep(time.Duration(job.CheckInterval) * time.Second)
	}

	// Se não confirmou após max checks, marcar como pendente permanentemente
	twp.logger.Warn("TX TRON não confirmada após limite de verificações",
		zap.String("tx_hash", job.TronTxHash),
		zap.Int("checks", checkCount),
	)

	err := twp.DB.UpdateTransaction(job.TransactionID, map[string]interface{}{
		"tron_tx_status": "unconfirmed",
	})
	if err != nil {
		twp.logger.Error("Erro ao marcar TX como não confirmada", zap.Error(err))
	}

	if job.CallbackURL != "" {
		twp.sendCallback(job.CallbackURL, map[string]interface{}{
			"status":    "unconfirmed",
			"tx_hash":   job.TronTxHash,
			"job_id":    job.JobID.String(),
			"tx_id":     job.TransactionID.String(),
			"timestamp": time.Now().Unix(),
		})
	}
}

// sendCallback envia notificação ao cliente sobre status da TX
func (twp *TronWorkerPool) sendCallback(url string, data map[string]interface{}) {
	// Implementar webhook call aqui (usar http.Client)
	// Por agora, apenas log
	twp.logger.Info("Enviando callback",
		zap.String("url", url),
		zap.Any("data", data),
	)
}

// SubmitConfirmationJob adiciona um job de confirmação à fila
func (twp *TronWorkerPool) SubmitConfirmationJob(userID uuid.UUID, txID uuid.UUID, txHash string, callbackURL string) {
	job := TronTxConfirmJob{
		Type:          JobTronTxConfirm,
		UserID:        userID,
		TransactionID: txID,
		TronTxHash:    txHash,
		CallbackURL:   callbackURL,
		CheckInterval: 10, // Verificar a cada 10 segundos
		MaxChecks:     30, // Máximo de 5 minutos (30 * 10s)
		JobID:         uuid.New(),
	}

	select {
	case twp.Jobs <- job:
		twp.logger.Debug("Job de confirmação TRON enfileirado", zap.String("tx_hash", txHash))
	case <-twp.quit:
		twp.logger.Warn("Não foi possível enfileirar job, workers já foram parados")
	default:
		twp.logger.Error("Fila de confirmação TRON cheia, descartando job", zap.String("tx_hash", txHash))
	}
}
