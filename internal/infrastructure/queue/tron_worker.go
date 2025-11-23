package workers

import (
	"bytes"
	"encoding/json"
	"financial-system-pro/internal/infrastructure/database"
	"fmt"
	"net/http"
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

	// Atualizar status para 'confirming' e enviar callback
	err := twp.DB.UpdateTransaction(job.TransactionID, map[string]interface{}{
		"tron_tx_status": "confirming",
	})
	if err != nil {
		twp.logger.Warn("failed to update status to confirming", zap.Error(err))
	}

	if job.CallbackURL != "" {
		twp.sendCallback(job.CallbackURL, map[string]interface{}{
			"status":       "confirming",
			"tx_hash":      job.TronTxHash,
			"job_id":       job.JobID.String(),
			"tx_id":        job.TransactionID.String(),
			"timestamp":    time.Now().Unix(),
			"check_count":  0,
			"max_checks":   job.MaxChecks,
			"description":  "Waiting for confirmations on TRON network",
			"explorer_url": fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", job.TronTxHash),
		})
	}

	// Verificar status da transação
	checkCount := 0
	for checkCount < job.MaxChecks {
		select {
		case <-twp.quit:
			twp.logger.Info("Parando monitoramento de TX", zap.String("tx_hash", job.TronTxHash))
			return
		default:
		}

		checkCount++

		// Buscar status da transação TRON
		_, err := twp.TronSvc.GetTransaction(job.TronTxHash)
		if err != nil {
			twp.logger.Warn("Erro ao buscar status de TX (ainda não confirmada)",
				zap.String("tx_hash", job.TronTxHash),
				zap.Error(err),
				zap.Int("check_count", checkCount),
			)

			// Aguardar antes de próxima tentativa
			time.Sleep(time.Duration(job.CheckInterval) * time.Second)
			continue
		}

		// Transação encontrada na blockchain
		// Primeira confirmação: status 'confirmed'
		err = twp.DB.UpdateTransaction(job.TransactionID, map[string]interface{}{
			"tron_tx_status": "confirmed",
			"tron_tx_hash":   job.TronTxHash,
		})
		if err != nil {
			twp.logger.Error("Erro ao atualizar status de TX no DB",
				zap.String("tx_hash", job.TronTxHash),
				zap.Error(err),
			)
		}

		twp.logger.Info("TX TRON confirmada (primeira confirmação)",
			zap.String("tx_hash", job.TronTxHash),
			zap.String("user_id", job.UserID.String()),
			zap.Int("check_count", checkCount),
		)

		// Enviar callback de 'confirmed'
		if job.CallbackURL != "" {
			twp.sendCallback(job.CallbackURL, map[string]interface{}{
				"status":        "confirmed",
				"tx_hash":       job.TronTxHash,
				"job_id":        job.JobID.String(),
				"tx_id":         job.TransactionID.String(),
				"timestamp":     time.Now().Unix(),
				"confirmations": 1,
				"description":   "Transaction confirmed on TRON network (1 confirmation)",
				"explorer_url":  fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", job.TronTxHash),
			})
		}

		// Aguardar mais confirmações antes de marcar como 'completed'
		time.Sleep(15 * time.Second)

		// Verificar novamente para ter múltiplas confirmações
		_, err = twp.TronSvc.GetTransaction(job.TronTxHash)
		if err == nil {
			// Atualizar para 'completed'
			err = twp.DB.UpdateTransaction(job.TransactionID, map[string]interface{}{
				"tron_tx_status": "completed",
			})
			if err != nil {
				twp.logger.Error("Erro ao atualizar status para completed",
					zap.String("tx_hash", job.TronTxHash),
					zap.Error(err),
				)
			}

			twp.logger.Info("TX TRON completed (múltiplas confirmações)",
				zap.String("tx_hash", job.TronTxHash),
				zap.String("user_id", job.UserID.String()),
			)

			// Enviar callback final de 'completed'
			if job.CallbackURL != "" {
				twp.sendCallback(job.CallbackURL, map[string]interface{}{
					"status":        "completed",
					"tx_hash":       job.TronTxHash,
					"job_id":        job.JobID.String(),
					"tx_id":         job.TransactionID.String(),
					"timestamp":     time.Now().Unix(),
					"confirmations": 3,
					"description":   "Transaction fully completed with multiple confirmations",
					"explorer_url":  fmt.Sprintf("https://shasta.tronscan.org/#/transaction/%s", job.TronTxHash),
				})
			}
		}

		return
	}

	// Se não confirmou após max checks, marcar como timeout
	twp.logger.Warn("TX TRON não confirmada após limite de verificações",
		zap.String("tx_hash", job.TronTxHash),
		zap.Int("checks", checkCount),
	)

	// Atualizar status no DB
	updateErr := twp.DB.UpdateTransaction(job.TransactionID, map[string]interface{}{
		"tron_tx_status": "timeout",
	})
	if updateErr != nil {
		twp.logger.Error("Erro ao atualizar status timeout no DB",
			zap.String("tx_hash", job.TronTxHash),
			zap.Error(updateErr),
		)
	}

	// Enviar callback de timeout
	if job.CallbackURL != "" {
		twp.sendCallback(job.CallbackURL, map[string]interface{}{
			"status":      "timeout",
			"tx_hash":     job.TronTxHash,
			"job_id":      job.JobID.String(),
			"tx_id":       job.TransactionID.String(),
			"timestamp":   time.Now().Unix(),
			"check_count": checkCount,
			"max_checks":  job.MaxChecks,
		})
	}
}

// sendCallback envia notificação HTTP ao cliente sobre status da TX
func (twp *TronWorkerPool) sendCallback(url string, data map[string]interface{}) {
	if url == "" {
		return
	}

	// Preparar payload JSON
	payload, err := json.Marshal(data)
	if err != nil {
		twp.logger.Error("Erro ao serializar callback payload",
			zap.String("url", url),
			zap.Error(err),
		)
		return
	}

	// Criar requisição HTTP POST
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		twp.logger.Error("Erro ao criar requisição de callback",
			zap.String("url", url),
			zap.Error(err),
		)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FinancialSystemPro-TronWorker/1.0")
	req.Header.Set("X-Callback-Event", "tron_transaction_update")

	// Enviar callback com timeout de 10 segundos
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		twp.logger.Warn("Erro ao enviar callback",
			zap.String("url", url),
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		twp.logger.Info("Callback enviado com sucesso",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
			zap.Any("data", data),
		)
	} else {
		twp.logger.Warn("Callback retornou status inesperado",
			zap.String("url", url),
			zap.Int("status_code", resp.StatusCode),
		)
	}
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
