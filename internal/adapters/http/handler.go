package http

import (
	"financial-system-pro/internal/application/dto"
	"financial-system-pro/internal/application/services"
	"financial-system-pro/internal/domain/entities"
	"financial-system-pro/internal/domain/errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// startTime registra quando a aplicação iniciou
var startTime = time.Now()

type Handler struct {
	userService        services.UserServiceInterface
	authService        services.AuthServiceInterface
	transactionService services.TransactionServiceInterface
	tronService        services.TronServiceInterface
	queueManager       QueueManagerInterface
	logger             LoggerInterface
	rateLimiter        RateLimiterInterface
}

// NewHandlerForTesting creates a handler instance for testing purposes
// This allows tests to inject mock services
func NewHandlerForTesting(
	userService services.UserServiceInterface,
	authService services.AuthServiceInterface,
	transactionService services.TransactionServiceInterface,
	tronService services.TronServiceInterface,
	queueManager QueueManagerInterface,
	logger LoggerInterface,
	rateLimiter RateLimiterInterface,
) *Handler {
	return &Handler{
		userService:        userService,
		authService:        authService,
		transactionService: transactionService,
		tronService:        tronService,
		queueManager:       queueManager,
		logger:             logger,
		rateLimiter:        rateLimiter,
	}
}

// handleAppError responde com status code e mensagem de erro do AppError
func (h *Handler) handleAppError(ctx *fiber.Ctx, err *errors.AppError) error {
	h.logger.Warn("API error",
		zap.String("code", err.Code),
		zap.String("message", err.Message),
		zap.Int("status", err.StatusCode),
	)
	return ctx.Status(err.StatusCode).JSON(err)
}

// checkDatabaseAvailable verifica se os serviços de banco estão disponíveis
func (h *Handler) CheckDatabaseAvailable(ctx *fiber.Ctx) error {
	if h.userService == nil || h.authService == nil || h.transactionService == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "Database connection not yet established",
			"message": "Please try again in a few moments",
		})
	}
	return nil
}

// CreateUser godoc
// @Summary      Cria um novo usuário
// @Description  Endpoint para criar usuário
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        userRequest  body  dto.UserRequest  true  "Dados do usuário"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/users [post]
func (h *Handler) CreateUser(ctx *fiber.Ctx) error {
	if err := h.CheckDatabaseAvailable(ctx); err != nil {
		return err
	}

	var userRequest dto.UserRequest
	if validErr := dto.ValidateRequest(ctx, &userRequest); validErr != nil {
		return h.handleAppError(ctx, validErr)
	}

	if appErr := h.userService.CreateNewUser(&userRequest); appErr != nil {
		return h.handleAppError(ctx, appErr)
	}

	h.logger.Info("user created successfully",
		zap.String("email", userRequest.Email),
	)
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created successfully"})
}

// CreateUser godoc
// @Summary      Autentica usuário
// @Description  Endpoint para autenticar usuário
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        loginRequest  body  dto.LoginRequest  true  "Dados de login"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/login [post]
func (h *Handler) Login(ctx *fiber.Ctx) error {
	if err := h.CheckDatabaseAvailable(ctx); err != nil {
		return err
	}

	var loginRequest dto.LoginRequest
	if validErr := dto.ValidateRequest(ctx, &loginRequest); validErr != nil {
		return h.handleAppError(ctx, validErr)
	}

	tokenJWT, appErr := h.authService.Login(&loginRequest)
	if appErr != nil {
		return h.handleAppError(ctx, appErr)
	}

	h.logger.Info("user logged in successfully",
		zap.String("email", loginRequest.Email),
	)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Login successful", "token": tokenJWT})
}

// CreateUser godoc
// @Summary      Deposita valor na conta do usuário
// @Description  Endpoint para depositar valor na conta do usuário
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        depositRequest  body  dto.DepositRequest  true  "Dados do depósito"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/deposit [post]
func (h *Handler) Deposit(ctx *fiber.Ctx) error {
	if err := h.CheckDatabaseAvailable(ctx); err != nil {
		return err
	}

	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}
	userID, ok := userIDLocal.(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var depositRequest dto.DepositRequest
	if validErr := dto.ValidateRequest(ctx, &depositRequest); validErr != nil {
		RecordFailure()
		return h.handleAppError(ctx, validErr)
	}

	amount, err := decimal.NewFromString(depositRequest.Amount)
	if err != nil {
		RecordFailure()
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		RecordFailure()
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	resp, err := h.transactionService.Deposit(userID, amount, depositRequest.CallbackURL)
	if err != nil {
		RecordFailure()
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	RecordDeposit()
	return ctx.Status(resp.StatusCode).JSON(resp.Body)
}

// CreateUser godoc
// @Summary      Consulta o saldo da conta do usuário
// @Description  Endpoint para consultar o saldo da conta do usuário
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/balance [get]
func (h *Handler) Balance(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userID, ok := userIDLocal.(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	balance, err := h.transactionService.GetBalance(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"balance": balance})
}

// GetUserWallet godoc
// @Summary      Retorna o endereço TRON da carteira do usuário
// @Description  Endpoint para obter o endereço TRON associado à conta do usuário autenticado
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}}
// @Failure      400  {object}  map[string]interface{}}
// @Failure      404  {object}  map[string]interface{}}
// @Failure      500  {object}  map[string]interface{}}
// @Router       /api/wallet [get]
func (h *Handler) GetUserWallet(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}

	userID, ok := userIDLocal.(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// Converter para UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	// Buscar wallet do usuário no BD
	walletInfo, err := h.transactionService.GetWalletInfo(uid)
	if err != nil {
		h.logger.Warn("wallet not found for user",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Wallet not found",
			"message": "User wallet has not been generated yet",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"wallet_address": walletInfo.TronAddress,
		"blockchain":     "tron",
		"user_id":        userID,
	})
}

// CreateUser godoc
// @Summary      Retira valor da conta do usuário
// @Description  Endpoint para retirar valor da conta do usuário. Para withdraw TRON, a wallet do usuário é usada automaticamente. Opcionalmente, um tron_address externo pode ser fornecido.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        withdrawRequest  body  dto.WithdrawRequest  true  "Dados do saque. withdraw_type='internal' debita saldo, withdraw_type='tron' envia da vault para carteira TRON do usuário"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}}
// @Failure      400  {object}  map[string]interface{}}
// @Failure      500  {object}  map[string]interface{}}
// @Router       /api/withdraw [post]
func (h *Handler) Withdraw(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}
	userID, ok := userIDLocal.(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var withdrawRequest dto.WithdrawRequest
	if validErr := dto.ValidateRequest(ctx, &withdrawRequest); validErr != nil {
		return h.handleAppError(ctx, validErr)
	}

	amount, err := decimal.NewFromString((withdrawRequest.Amount))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	withdrawType := withdrawRequest.WithdrawType
	if withdrawType == "" {
		withdrawType = "internal"
	}

	var resp *services.ServiceResponse
	if withdrawType == "tron" {
		resp, err = h.transactionService.WithdrawTron(userID, amount, withdrawRequest.CallbackURL)
	} else {
		resp, err = h.transactionService.Withdraw(userID, amount, withdrawRequest.CallbackURL)
	}

	if err != nil {
		RecordFailure()
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusAccepted {
		RecordWithdraw()
	}

	return ctx.Status(resp.StatusCode).JSON(resp.Body)
}

// CreateUser godoc
// @Summary      Transferred valor para outra conta
// @Description  Endpoint para transferir valor para outra conta
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        transferRequest  body  dto.TransferRequest  true  "Dados da transferência"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/transfer [post]
func (h *Handler) Transfer(ctx *fiber.Ctx) error {
	userIDLocal := ctx.Locals("user_id")
	if userIDLocal == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id not found"})
	}
	userID, ok := userIDLocal.(string)
	if !ok || userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var transferRequest dto.TransferRequest
	if validErr := dto.ValidateRequest(ctx, &transferRequest); validErr != nil {
		return h.handleAppError(ctx, validErr)
	}

	amount, err := decimal.NewFromString(transferRequest.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	resp, err := h.transactionService.Transfer(userID, amount, transferRequest.To, transferRequest.CallbackURL)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(resp.StatusCode).JSON(resp.Body)
}

// GetTronBalance godoc
// @Summary      Obtém saldo de uma carteira Tron
// @Description  Endpoint para obter saldo de uma carteira Tron Testnet
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Param        address  query  string  true  "Endereço Tron (começa com T)"
// @Security     BearerAuth
// @Success      200  {object}  dto.TronBalance
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/balance [get]
func (h *Handler) GetTronBalance(ctx *fiber.Ctx) error {
	address := ctx.Query("address")
	if address == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "address is required"})
	}

	if !h.tronService.ValidateAddress(address) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid Tron address"})
	}

	balance, err := h.tronService.GetBalance(address)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"address":     address,
		"balance_sun": balance,
		"balance_trx": float64(balance) / 1000000,
	})
}

// SendTronTransaction godoc
// @Summary      Envia uma transação Tron
// @Description  Endpoint para enviar TRX na rede Tron Testnet
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Param        txRequest  body  entities.TronTransactionRequest  true  "Dados da transação"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/send [post]
func (h *Handler) SendTronTransaction(ctx *fiber.Ctx) error {
	var txRequest entities.TronTransactionRequest
	if validErr := dto.ValidateRequest(ctx, &txRequest); validErr != nil {
		return h.handleAppError(ctx, validErr)
	}

	if !h.tronService.ValidateAddress(txRequest.FromAddress) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid from address"})
	}
	if !h.tronService.ValidateAddress(txRequest.ToAddress) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid to address"})
	}

	txHash, err := h.tronService.SendTransaction(
		txRequest.FromAddress,
		txRequest.ToAddress,
		txRequest.Amount,
		txRequest.PrivateKey,
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Transaction sent successfully",
		"tx_hash": txHash,
		"from":    txRequest.FromAddress,
		"to":      txRequest.ToAddress,
		"amount":  txRequest.Amount,
	})
}

// GetTronTransactionStatus godoc
// @Summary      Obtém status de uma transação Tron
// @Description  Endpoint para verificar o status de uma transação Tron
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Param        tx_hash  query  string  true  "Hash da transação"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/tx-status [get]
func (h *Handler) GetTronTransactionStatus(ctx *fiber.Ctx) error {
	txHash := ctx.Query("tx_hash")
	if txHash == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tx_hash is required"})
	}

	status, err := h.tronService.GetTransactionStatus(txHash)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"tx_hash": txHash,
		"status":  status,
	})
}

// CreateTronWallet godoc
// @Summary      Cria uma nova carteira Tron
// @Description  Endpoint para criar uma nova carteira Tron Testnet
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201  {object}  entities.TronWallet
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/wallet [post]
func (h *Handler) CreateTronWallet(ctx *fiber.Ctx) error {
	wallet, err := h.tronService.CreateWallet()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(wallet)
}

// CheckTronNetwork godoc
// @Summary      Verifica conexão com rede Tron
// @Description  Endpoint para verificar se está conectado à rede Tron Testnet
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/network [get]
func (h *Handler) CheckTronNetwork(ctx *fiber.Ctx) error {
	isConnected := h.tronService.IsTestnetConnected()

	if !isConnected {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "disconnected",
			"message": "Não foi possível conectar à rede Tron Testnet",
		})
	}

	networkInfo, err := h.tronService.GetNetworkInfo()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":       "connected",
		"network":      "Tron Testnet (Shasta)",
		"network_info": networkInfo,
	})
}

// EstimateTronGas godoc
// @Summary      Estima energia necessária para transação
// @Description  Endpoint para estimar o gasto de energia para uma transação Tron
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Param        estimateRequest  body  map[string]interface{}  true  "from_address, to_address, amount"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/estimate-energy [post]
func (h *Handler) EstimateTronGas(ctx *fiber.Ctx) error {
	var estimateReq struct {
		FromAddress string `json:"from_address" validate:"required"`
		ToAddress   string `json:"to_address" validate:"required"`
		Amount      int64  `json:"amount" validate:"required,gt=0"`
	}

	if err := ctx.BodyParser(&estimateReq); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if !h.tronService.ValidateAddress(estimateReq.FromAddress) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid from address"})
	}
	if !h.tronService.ValidateAddress(estimateReq.ToAddress) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid to address"})
	}

	estimatedEnergy, err := h.tronService.EstimateGasForTransaction(
		estimateReq.FromAddress,
		estimateReq.ToAddress,
		estimateReq.Amount,
	)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"from_address":     estimateReq.FromAddress,
		"to_address":       estimateReq.ToAddress,
		"amount":           estimateReq.Amount,
		"estimated_energy": estimatedEnergy,
		"message":          "Estimativa de energia para transação",
	})
}

// GetRPCStatus godoc
// @Summary      Obtém status das conexões RPC
// @Description  Endpoint para verificar status do RPC e gRPC
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/rpc-status [get]
func (h *Handler) GetRPCStatus(ctx *fiber.Ctx) error {
	status := h.tronService.GetConnectionStatus()

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    status,
		"timestamp": time.Now().Unix(),
	})
}

// GetAvailableMethods godoc
// @Summary      Lista métodos RPC disponíveis
// @Description  Endpoint que retorna os métodos RPC disponíveis para interagir com Tron
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/rpc-methods [get]
func (h *Handler) GetAvailableMethods(ctx *fiber.Ctx) error {
	methods := entities.GetAvailableTronRPCMethods()

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"available_methods": methods,
		"total_methods":     len(methods),
		"documentation":     "https://tronprotocol.github.io/documentation-en/",
	})
}

// CallRPCMethod godoc
// @Summary      Chama um método RPC customizado
// @Description  Endpoint para chamar qualquer método RPC disponível no Tron
// @Tags         Tron
// @Accept       json
// @Produce      json
// @Param        rpcCall  body  map[string]interface{}  true  "method, params"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/rpc-call [post]
func (h *Handler) CallRPCMethod(ctx *fiber.Ctx) error {
	var rpcCall struct {
		Method string        `json:"method" validate:"required"`
		Params []interface{} `json:"params"`
	}

	if err := ctx.BodyParser(&rpcCall); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	rpcClient := h.tronService.GetRPCClient()
	if rpcClient == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "RPC client not available",
		})
	}

	result, err := rpcClient.Call(ctx.Context(), rpcCall.Method, rpcCall.Params...)
	if err != nil {
		h.tronService.RecordError(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"method": rpcCall.Method,
		"params": rpcCall.Params,
		"result": result,
	})
}

// TestQueueDeposit testa o envio de uma tarefa de depósito à fila
// POST /api/queue/test-deposit
// Body: {"user_id": "uuid", "amount": "100.00"}
func (h *Handler) TestQueueDeposit(ctx *fiber.Ctx) error {
	if h.queueManager == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Queue manager not initialized",
		})
	}

	var req struct {
		UserID string `json:"user_id"`
		Amount string `json:"amount"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.Amount == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and amount are required",
		})
	}

	taskID, err := h.queueManager.EnqueueDeposit(ctx.Context(), req.UserID, req.Amount, "")
	if err != nil {
		h.logger.Error("failed to enqueue deposit", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to enqueue task",
		})
	}

	h.logger.Info("test deposit task enqueued",
		zap.String("task_id", taskID),
		zap.String("user_id", req.UserID),
	)

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"task_id": taskID,
		"status":  "queued",
		"message": "Deposit task has been queued for processing",
	})
}

// GetAuditLogs retorna logs de auditoria (placeholder - delegating to actual implementation)
func (h *Handler) GetAuditLogs(ctx *fiber.Ctx) error {
	// Note: This is a placeholder. In production, this would need:
	// 1. Database access for audit logs
	// 2. Proper authorization (admin only)
	// 3. Pagination and filtering
	h.logger.Info("audit logs endpoint called")

	return ctx.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   "Audit logs endpoint requires database integration",
		"message": "Use database.GetAuditLogs() method directly or implement full AuditHandler",
	})
}

// GetAuditStats retorna estatísticas de auditoria (placeholder)
func (h *Handler) GetAuditStats(ctx *fiber.Ctx) error {
	h.logger.Info("audit stats endpoint called")

	return ctx.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   "Audit stats endpoint requires implementation",
		"message": "Future feature: statistics and analytics for audit logs",
	})
}
