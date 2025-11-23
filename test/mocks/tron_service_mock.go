package mocks

import (
	"context"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

// MockTronService is a mock implementation of TronService
type MockTronService struct {
	mock.Mock
	IsConnected    bool
	ValidAddresses map[string]bool
}

func NewMockTronService() *MockTronService {
	return &MockTronService{
		IsConnected:    true,
		ValidAddresses: make(map[string]bool),
	}
}

func (m *MockTronService) GetBalance(address string) (int64, error) {
	if m.ValidAddresses != nil && !m.ValidAddresses[address] {
		return 0, nil
	}
	return 1000000, nil
}

func (m *MockTronService) ValidateAddress(address string) bool {
	if m.ValidAddresses != nil {
		return m.ValidAddresses[address]
	}
	return false
}

func (m *MockTronService) SendTransaction(from, to string, amount int64, privateKey string) (string, error) {
	return "mocktxhash123", nil
}

func (m *MockTronService) GetTransactionStatus(txHash string) (string, error) {
	return "confirmed", nil
}

func (m *MockTronService) CreateWallet() (*entities.TronWallet, error) {
	return &entities.TronWallet{
		Address:    "TMockAddress123",
		PrivateKey: "mockprivatekey",
		PublicKey:  "mockpublickey",
	}, nil
}

func (m *MockTronService) IsTestnetConnected() bool {
	return m.IsConnected
}

func (m *MockTronService) GetNetworkInfo() (map[string]interface{}, error) {
	return map[string]interface{}{
		"network": "testnet",
		"chain":   "tron",
	}, nil
}

func (m *MockTronService) EstimateGasForTransaction(from, to string, amount int64) (int64, error) {
	return 15000, nil
}

func (m *MockTronService) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{
		"rpc_connected":  m.IsConnected,
		"grpc_connected": m.IsConnected,
	}
}

func (m *MockTronService) GetRPCClient() *services.RPCClient {
	return nil
}

func (m *MockTronService) RecordError(err error) {
	// mock implementation
}

func (m *MockTronService) HealthCheck(ctx context.Context) error {
	if !m.IsConnected {
		return context.DeadlineExceeded
	}
	return nil
}

func (m *MockTronService) HasVaultConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTronService) GetVaultAddress() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTronService) GetVaultPrivateKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTronService) GetTransactionInfo(txHash string) (map[string]interface{}, error) {
	args := m.Called(txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
