package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Métricas por Bounded Context

// Transaction Context Metrics
var (
	TransactionDepositTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_deposit_total",
			Help: "Total number of deposit transactions",
		},
		[]string{"status"}, // status: success, failed
	)

	TransactionWithdrawTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_withdraw_total",
			Help: "Total number of withdraw transactions",
		},
		[]string{"status"},
	)

	TransactionTransferTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_transfer_total",
			Help: "Total number of transfer transactions",
		},
		[]string{"status"},
	)

	TransactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_amount_dollars",
			Help:    "Distribution of transaction amounts in dollars",
			Buckets: []float64{1, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"type"}, // type: deposit, withdraw, transfer
	)

	TransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_duration_seconds",
			Help:    "Duration of transaction processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type", "status"},
	)

	TransactionQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "transaction_queue_size",
			Help: "Current size of transaction queue",
		},
	)
)

// User Context Metrics
var (
	UserCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_created_total",
			Help: "Total number of users created",
		},
	)

	UserAuthenticationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_authentication_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"status"}, // status: success, failed
	)

	UserActiveTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "user_active_total",
			Help: "Current number of active users",
		},
	)

	UserBalanceTotal = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "user_balance_dollars",
			Help:    "Distribution of user balances",
			Buckets: []float64{0, 10, 100, 1000, 10000, 100000},
		},
	)
)

// Blockchain Context Metrics
var (
	BlockchainWalletCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "blockchain_wallet_created_total",
			Help: "Total number of blockchain wallets created",
		},
		[]string{"blockchain_type"}, // blockchain_type: TRON, ETH, etc
	)

	BlockchainTransactionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "blockchain_transaction_total",
			Help: "Total number of blockchain transactions",
		},
		[]string{"blockchain_type", "status"},
	)

	BlockchainTransactionConfirmations = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "blockchain_transaction_confirmations",
			Help:    "Number of confirmations for blockchain transactions",
			Buckets: []float64{1, 3, 6, 12, 24, 48},
		},
		[]string{"blockchain_type"},
	)

	BlockchainRPCCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "blockchain_rpc_call_duration_seconds",
			Help:    "Duration of RPC calls to blockchain",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"blockchain_type", "method", "status"},
	)

	BlockchainConnectionStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "blockchain_connection_status",
			Help: "Blockchain connection status (1 = connected, 0 = disconnected)",
		},
		[]string{"blockchain_type"},
	)
)

// Queue Context Metrics
var (
	QueueJobsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_jobs_processed_total",
			Help: "Total number of queue jobs processed",
		},
		[]string{"queue_name", "status"},
	)

	QueueJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "queue_job_duration_seconds",
			Help:    "Duration of queue job processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue_name"},
	)

	QueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_size",
			Help: "Current size of queue",
		},
		[]string{"queue_name"},
	)

	QueueFailedJobs = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_failed_jobs_total",
			Help: "Total number of failed queue jobs",
		},
		[]string{"queue_name", "error_type"},
	)
)

// System Metrics
var (
	SystemHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_health_status",
			Help: "System health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"component"},
	)

	SystemUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_uptime_seconds",
			Help: "System uptime in seconds",
		},
	)
)

// Helper functions para registrar métricas facilmente

// RecordDeposit registra uma métrica de depósito
func RecordDeposit(amountUSD float64, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	TransactionDepositTotal.WithLabelValues(status).Inc()
	if success {
		TransactionAmount.WithLabelValues("deposit").Observe(amountUSD)
	}
}

// RecordWithdraw registra uma métrica de saque
func RecordWithdraw(amountUSD float64, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	TransactionWithdrawTotal.WithLabelValues(status).Inc()
	if success {
		TransactionAmount.WithLabelValues("withdraw").Observe(amountUSD)
	}
}

// RecordTransfer registra uma métrica de transferência
func RecordTransfer(amountUSD float64, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	TransactionTransferTotal.WithLabelValues(status).Inc()
	if success {
		TransactionAmount.WithLabelValues("transfer").Observe(amountUSD)
	}
}

// RecordUserCreated registra criação de usuário
func RecordUserCreated() {
	UserCreatedTotal.Inc()
}

// RecordAuthentication registra tentativa de autenticação
func RecordAuthentication(success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	UserAuthenticationTotal.WithLabelValues(status).Inc()
}

// RecordWalletCreated registra criação de carteira
func RecordWalletCreated(blockchainType string) {
	BlockchainWalletCreatedTotal.WithLabelValues(blockchainType).Inc()
}

// RecordBlockchainTransaction registra transação blockchain
func RecordBlockchainTransaction(blockchainType string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	BlockchainTransactionTotal.WithLabelValues(blockchainType, status).Inc()
}
