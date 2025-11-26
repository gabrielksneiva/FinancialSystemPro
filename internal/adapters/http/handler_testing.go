package http

import (
	"context"
	"encoding/json"
	services "financial-system-pro/internal/application/services"
	"financial-system-pro/internal/contexts/transaction/application/service"
	userDDD "financial-system-pro/internal/contexts/user/application/service"
	"financial-system-pro/internal/domain/entities"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var startTime = time.Now()

// Handler (test-only minimal) providing Tron endpoints used by tron_endpoints_test.go
type Handler struct {
	dddUserService        *userDDD.UserService
	dddTransactionService *service.TransactionService
	// legacy shim fields for older tests
	transactionService interface{}
	tronGateway        interface { // narrowed to concrete TronGateway extended surface
		ValidateAddress(string) bool
		GetBalance(string) (int64, error)
		SendTransaction(string, string, int64, string) (string, error)
		GetTransactionStatus(string) (string, error)
		CreateWallet() (*entities.TronWallet, error)
		IsTestnetConnected() bool
		GetNetworkInfo() (map[string]interface{}, error)
		EstimateGasForTransaction(string, string, int64) (int64, error)
		GetRPCClient() *services.RPCClient
		GetConnectionStatus() map[string]interface{}
	}
	queueManager QueueManagerInterface
	// accept any logger for test flexibility (zap.Logger or adapter)
	logger      interface{}
	rateLimiter RateLimiterInterface
}

// NewHandlerForTesting builds a minimal Handler for HTTP tests.
func NewHandlerForTesting(
	user interface{},
	tx interface{},
	tron interface{}, // expect *gateway.TronGateway or mock implementing same methods
	queue interface{},
	logger interface{},
	rl RateLimiterInterface,
) *Handler {
	var qm QueueManagerInterface
	if queue != nil {
		if q, ok := queue.(QueueManagerInterface); ok {
			qm = q
		}
	}

	// Safely assert tron service to expected interface (tests provide a mock)
	var ts interface {
		ValidateAddress(string) bool
		GetBalance(string) (int64, error)
		SendTransaction(string, string, int64, string) (string, error)
		GetTransactionStatus(string) (string, error)
		CreateWallet() (*entities.TronWallet, error)
		IsTestnetConnected() bool
		GetNetworkInfo() (map[string]interface{}, error)
		EstimateGasForTransaction(string, string, int64) (int64, error)
		GetRPCClient() *services.RPCClient
		GetConnectionStatus() map[string]interface{}
	}
	if tron != nil {
		if t, ok := tron.(interface {
			ValidateAddress(string) bool
			GetBalance(string) (int64, error)
			SendTransaction(string, string, int64, string) (string, error)
			GetTransactionStatus(string) (string, error)
			CreateWallet() (*entities.TronWallet, error)
			IsTestnetConnected() bool
			GetNetworkInfo() (map[string]interface{}, error)
			EstimateGasForTransaction(string, string, int64) (int64, error)
			GetRPCClient() *services.RPCClient
			GetConnectionStatus() map[string]interface{}
		}); ok {
			ts = t
		}
	}
	var dddUser *userDDD.UserService
	if u, ok := user.(*userDDD.UserService); ok {
		dddUser = u
	}
	var dddTx *service.TransactionService
	if t, ok := tx.(*service.TransactionService); ok {
		dddTx = t
	}
	return &Handler{
		dddUserService:        dddUser,
		dddTransactionService: dddTx,
		transactionService:    tx,
		tronGateway:           ts,
		queueManager:          qm,
		logger:                logger,
		rateLimiter:           rl,
	}
}

// newMockLogger provides a basic logger for tests.
func newMockLogger() LoggerInterface { return NewZapLoggerAdapter(zap.NewNop()) }

// WithMultiChainWalletService sets multichain wallet service
// Legacy compatibility: accept MultiChainWalletService but do nothing
func (h *Handler) WithMultiChainWalletService(_ *services.MultiChainWalletService) *Handler { return h }

// --- Legacy handler stubs used by router.go ---
func (h *Handler) TestQueueDeposit(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) }
func (h *Handler) CreateUser(c *fiber.Ctx) error       { return c.SendStatus(fiber.StatusCreated) }
func (h *Handler) Login(c *fiber.Ctx) error            { return c.SendStatus(fiber.StatusOK) }
func (h *Handler) GetAuditLogs(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented"})
}
func (h *Handler) GetAuditStats(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented"})
}
func (h *Handler) Deposit(c *fiber.Ctx) error  { return c.SendStatus(fiber.StatusAccepted) }
func (h *Handler) Withdraw(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusAccepted) }
func (h *Handler) Transfer(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusAccepted) }
func (h *Handler) Balance(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"balance": 0})
}
func (h *Handler) GetUserWallet(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"wallet": fiber.Map{"address": "ADDR"}})
}
func (h *Handler) GenerateWallet(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"address": "ADDR"})
}

// GetTronBalance handles GET /api/tron/balance
func (h *Handler) GetTronBalance(c *fiber.Ctx) error {
	if h.tronGateway == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "tron service not available"})
	}
	addr := c.Query("address")
	if addr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "address is required"})
	}
	if !h.tronGateway.ValidateAddress(addr) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid Tron address"})
	}
	bal, err := h.tronGateway.GetBalance(addr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	// TRX = SUN / 1_000_000
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"balance_sun": bal, "balance_trx": float64(bal) / 1_000_000})
}

// SendTronTransaction handles POST /api/tron/send
func (h *Handler) SendTronTransaction(c *fiber.Ctx) error {
	if h.tronGateway == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "tron service not available"})
	}
	var req struct {
		From    string `json:"from_address"`
		To      string `json:"to_address"`
		Private string `json:"private_key"`
		Amount  int64  `json:"amount"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if !h.tronGateway.ValidateAddress(req.From) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid from address"})
	}
	if !h.tronGateway.ValidateAddress(req.To) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid to address"})
	}
	hash, err := h.tronGateway.SendTransaction(req.From, req.To, req.Amount, req.Private)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"tx_hash": hash})
}

// GetTronTransactionStatus handles GET /api/tron/tx-status
func (h *Handler) GetTronTransactionStatus(c *fiber.Ctx) error {
	if h.tronGateway == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "tron service not available"})
	}
	hash := c.Query("tx_hash")
	if hash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tx_hash is required"})
	}
	status, err := h.tronGateway.GetTransactionStatus(hash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": status})
}

// CreateTronWallet handles POST /api/tron/wallet
func (h *Handler) CreateTronWallet(c *fiber.Ctx) error {
	if h.tronGateway == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "tron service not available"})
	}
	w, err := h.tronGateway.CreateWallet()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"address": w.Address, "public_key": w.PublicKey})
}

// CheckTronNetwork handles GET /api/tron/network
func (h *Handler) CheckTronNetwork(c *fiber.Ctx) error {
	if h.tronGateway == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "tron service not available"})
	}
	if !h.tronGateway.IsTestnetConnected() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "disconnected"})
	}
	info, err := h.tronGateway.GetNetworkInfo()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(info)
}

// EstimateTronGas handles POST /api/tron/estimate-energy
func (h *Handler) EstimateTronGas(c *fiber.Ctx) error {
	if h.tronGateway == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "tron service not available"})
	}
	var req struct {
		From   string `json:"from_address"`
		To     string `json:"to_address"`
		Amount int64  `json:"amount"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if !h.tronGateway.ValidateAddress(req.From) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid from address"})
	}
	if !h.tronGateway.ValidateAddress(req.To) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid to address"})
	}
	gas, err := h.tronGateway.EstimateGasForTransaction(req.From, req.To, req.Amount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"estimated_energy": gas})
}

// GetRPCStatus handles GET /api/tron/rpc-status
func (h *Handler) GetRPCStatus(c *fiber.Ctx) error {
	status := h.tronGateway.GetConnectionStatus()
	return c.Status(fiber.StatusOK).JSON(status)
}

// GetAvailableMethods handles GET /api/tron/rpc-methods
func (h *Handler) GetAvailableMethods(c *fiber.Ctx) error {
	// Return a static list for testing, regardless of RPC availability
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"available_methods": []string{"call", "ping"}})
}

// CallRPCMethod handles POST /api/tron/rpc-call
func (h *Handler) CallRPCMethod(c *fiber.Ctx) error {
	var req struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	client := h.tronGateway.GetRPCClient()
	if client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "rpc not available"})
	}
	if req.Method == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "method is required"})
	}
	// forward call to RPC client and surface errors
	res, err := client.Call(context.Background(), req.Method)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	// echo raw result
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"result": res})
}
