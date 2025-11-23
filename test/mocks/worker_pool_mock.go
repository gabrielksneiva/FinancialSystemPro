package mocks

import (
	queue "financial-system-pro/internal/infrastructure/queue"

	"github.com/stretchr/testify/mock"
)

// MockWorkerPool is a mock implementation of worker pool
type MockWorkerPool struct {
	Jobs chan queue.TransactionJob
	mock.Mock
}

func NewMockWorkerPool() *MockWorkerPool {
	return &MockWorkerPool{
		Jobs: make(chan queue.TransactionJob, 100),
	}
}

func (m *MockWorkerPool) Start() {
	m.Called()
}

func (m *MockWorkerPool) Stop() {
	m.Called()
}

func (m *MockWorkerPool) SubmitJob(job queue.TransactionJob) error {
	args := m.Called(job)
	return args.Error(0)
}

// MockTronWorkerPool is a mock implementation of TRON worker pool
type MockTronWorkerPool struct {
	Jobs chan queue.TronTxConfirmJob
	mock.Mock
}

func NewMockTronWorkerPool() *MockTronWorkerPool {
	return &MockTronWorkerPool{
		Jobs: make(chan queue.TronTxConfirmJob, 100),
	}
}

func (m *MockTronWorkerPool) Start() {
	m.Called()
}

func (m *MockTronWorkerPool) Stop() {
	m.Called()
}

func (m *MockTronWorkerPool) SubmitConfirmationJob(job queue.TronTxConfirmJob) {
	m.Called(job)
	m.Jobs <- job
}
