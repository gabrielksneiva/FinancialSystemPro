package workers

import (
	"bytes"
	"encoding/json"
	"financial-system-pro/repositories"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var sugar = logger.Sugar()

func NewTransactionWorkerPool(db *repositories.NewDatabase, workers int, queueSize int) *TransactionWorkerPool {
	p := &TransactionWorkerPool{
		DB:      db,
		Jobs:    make(chan TransactionJob, queueSize),
		Workers: workers,
		quit:    make(chan struct{}),
	}

	for i := 0; i < p.Workers; i++ {
		go p.worker(i)
	}

	return p
}

func (p *TransactionWorkerPool) Stop() {
	close(p.quit)
	close(p.Jobs)
}

func (p *TransactionWorkerPool) worker(id int) {
	for {
		select {
		case job, ok := <-p.Jobs:
			if !ok {
				sugar.Infof("[Worker %d] job channel fechado, encerrando...", id)
				return
			}

			queueLen := len(p.Jobs)
			sugar.Infow("Job recebido",
				"worker_id", id,
				"job_type", job.Type,
				"account", job.Account.String(),
				"queue_size", queueLen,
				"callback_url", job.CallbackURL,
			)

			// simulação de "demora" p/ visualizar a fila
			time.Sleep(1 * time.Second)

			var err error
			switch job.Type {
			case JobDeposit:
				err = p.handleDeposit(job.Account, job.Amount)
			case JobWithdraw:
				err = p.handleWithdraw(job.Account, job.Amount)
			case JobTransfer:
				err = p.handleTransfer(job)
			}

			status := "done"
			if err != nil {
				status = "failed"
				sugar.Errorw("Erro ao processar job",
					"worker_id", id,
					"job_type", job.Type,
					"account", job.Account.String(),
					"error", err,
				)
			} else {
				sugar.Infow("Job concluído",
					"worker_id", id,
					"job_type", job.Type,
					"account", job.Account.String(),
					"status", status,
				)
			}

			if job.CallbackURL != "" {
				result := JobResult{
					JobID:   job.JobID.String(),
					JobType: string(job.Type),
					Account: job.Account.String(),
					Status:  status,
				}
				if err != nil {
					result.Error = err.Error()
				}

				go sendCallback(job.CallbackURL, result)
				sugar.Infow("Callback enviado",
					"worker_id", id,
					"job_id", job.JobID.String(),
					"url", job.CallbackURL,
				)
			}

		case <-p.quit:
			sugar.Infof("[Worker %d] encerrando loop por quit signal", id)
			return
		}
	}
}

func (p *TransactionWorkerPool) handleDeposit(account uuid.UUID, amount decimal.Decimal) error {
	if err := p.DB.Transaction(account, amount, "deposit"); err != nil {
		return err
	}
	return p.DB.Insert(&repositories.Transaction{
		AccountID:   account,
		Amount:      amount,
		Type:        "deposit",
		Category:    "credit",
		Description: "User deposit",
	})
}

func (p *TransactionWorkerPool) handleWithdraw(account uuid.UUID, amount decimal.Decimal) error {
	if err := p.DB.Transaction(account, amount, "withdraw"); err != nil {
		return err
	}
	return p.DB.Insert(&repositories.Transaction{
		AccountID:   account,
		Amount:      amount,
		Type:        "withdraw",
		Category:    "debit",
		Description: "User withdraw",
	})
}

func (p *TransactionWorkerPool) handleTransfer(job TransactionJob) error {
	amount := job.Amount
	userFrom := job.Account

	foundUser, err := p.DB.FindUserByField("email", job.ToEmail)
	if err != nil {
		return err
	}
	userTo := foundUser.ID

	foundUserFrom, err := p.DB.FindUserByField("id", userFrom.String())
	if err != nil {
		return err
	}

	if err := p.DB.Transaction(userFrom, amount, "withdraw"); err != nil {
		return err
	}
	if err := p.DB.Transaction(userTo, amount, "deposit"); err != nil {
		return err
	}

	if err := p.DB.Insert(&repositories.Transaction{
		AccountID:   userFrom,
		Amount:      amount,
		Type:        "transfer",
		Category:    "debit",
		Description: "User transfer to " + job.ToEmail,
	}); err != nil {
		return err
	}

	return p.DB.Insert(&repositories.Transaction{
		AccountID:   userTo,
		Amount:      amount,
		Type:        "transfer",
		Category:    "credit",
		Description: "User transfer from " + foundUserFrom.Email,
	})
}

func sendCallback(url string, result JobResult) {
	body, _ := json.Marshal(result)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("[Worker] failed to POST callback to %s: %v\n", url, err)
	} else {
		fmt.Printf("[Worker] sent callback to %s\n", url)
	}
}
