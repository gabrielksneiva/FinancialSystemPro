package dto

// HealthStatus representa o status de cada componente
type HealthStatus struct {
	Status   string            `json:"status"`
	Uptime   int64             `json:"uptime_seconds"`
	Services map[string]string `json:"services"`
}

// ProbeResponse resposta de probe
type ProbeResponse struct {
	Ready  bool   `json:"ready,omitempty"`
	Alive  bool   `json:"alive,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// MetricsResponse resposta com métricas
type MetricsResponse struct {
	Transactions struct {
		Deposits  int64 `json:"deposits"`
		Withdraws int64 `json:"withdraws"`
		Transfers int64 `json:"transfers"`
		Failures  int64 `json:"failures"`
		Total     int64 `json:"total"`
	} `json:"transactions"`
	API struct {
		TotalRequests     int64   `json:"total_requests"`
		AvgResponseTimeMs float64 `json:"avg_response_time_ms"`
	} `json:"api"`
	System struct {
		UptimeSeconds int64  `json:"uptime_seconds"`
		MemoryMb      uint64 `json:"memory_mb"`
		Goroutines    int    `json:"goroutines"`
		GCRuns        uint32 `json:"gc_runs"`
	} `json:"system"`
}
