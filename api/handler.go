package api

import (
	"financial-system-pro/domain"
	"financial-system-pro/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

type NewHandler struct {
	userService        *services.NewUserService
	authService        *services.NewAuthService
	transactionService *services.NewTransactionService
	tronService        *services.TronService
}

// checkDatabaseAvailable verifica se os serviços de banco estão disponíveis
func (h *NewHandler) checkDatabaseAvailable(ctx *fiber.Ctx) error {
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
// @Param        userRequest  body  domain.UserRequest  true  "Dados do usuário"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/users [post]
func (h *NewHandler) CreateUser(ctx *fiber.Ctx) error {
	if err := h.checkDatabaseAvailable(ctx); err != nil {
		return err
	}

	var userRequest domain.UserRequest
	err := isValid(ctx, &userRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err = h.userService.CreateNewUser(&userRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created succesfully"})
}

// CreateUser godoc
// @Summary      Autentica usuário
// @Description  Endpoint para autenticar usuário
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        loginRequest  body  domain.LoginRequest  true  "Dados de login"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/login [post]
func (h *NewHandler) Login(ctx *fiber.Ctx) error {
	if err := h.checkDatabaseAvailable(ctx); err != nil {
		return err
	}

	var loginRequest domain.LoginRequest
	err := isValid(ctx, &loginRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenJWT, err := h.authService.Login(&loginRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Login succesfully", "token": tokenJWT})
}

// CreateUser godoc
// @Summary      Deposita valor na conta do usuário
// @Description  Endpoint para depositar valor na conta do usuário
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        depositRequest  body  domain.DepositRequest  true  "Dados do depósito"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/deposit [post]
func (h *NewHandler) Deposit(ctx *fiber.Ctx) error {
	if err := h.checkDatabaseAvailable(ctx); err != nil {
		return err
	}

	var depositRequest domain.DepositRequest
	err := isValid(ctx, &depositRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	amount, err := decimal.NewFromString(depositRequest.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	resp, err := h.transactionService.Deposit(ctx, amount, depositRequest.CallbackURL)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

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
func (h *NewHandler) Balance(ctx *fiber.Ctx) error {
	UserID := ctx.Locals("ID").(string)
	if UserID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	balance, err := h.transactionService.GetBalance(ctx, UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"balance": balance})
}

// CreateUser godoc
// @Summary      Retira valor da conta do usuário
// @Description  Endpoint para retirar valor da conta do usuário
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        withdrawRequest  body  domain.WithdrawRequest  true  "Dados do saque"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/withdraw [post]
func (h *NewHandler) Withdraw(ctx *fiber.Ctx) error {
	var withdrawRequest domain.WithdrawRequest
	err := isValid(ctx, &withdrawRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	amount, err := decimal.NewFromString((withdrawRequest.Amount))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	resp, err := h.transactionService.Withdraw(ctx, amount, withdrawRequest.CallbackURL)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(resp.StatusCode).JSON(resp.Body)
}

// CreateUser godoc
// @Summary      Transfere valor para outra conta
// @Description  Endpoint para transferir valor para outra conta
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        transferRequest  body  domain.TransferRequest  true  "Dados da transferência"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/transfer [post]
func (h *NewHandler) Transfer(ctx *fiber.Ctx) error {
	var transferRequest domain.TransferRequest
	err := isValid(ctx, &transferRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	amount, err := decimal.NewFromString(transferRequest.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid amount format"})
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Amount must be greater than zero"})
	}

	resp, err := h.transactionService.Transfer(ctx, amount, transferRequest.To, transferRequest.CallbackURL)
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
// @Success      200  {object}  domain.TronBalance
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/balance [get]
func (h *NewHandler) GetTronBalance(ctx *fiber.Ctx) error {
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
// @Param        txRequest  body  domain.TronTransactionRequest  true  "Dados da transação"
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/send [post]
func (h *NewHandler) SendTronTransaction(ctx *fiber.Ctx) error {
	var txRequest domain.TronTransactionRequest
	err := isValid(ctx, &txRequest)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
func (h *NewHandler) GetTronTransactionStatus(ctx *fiber.Ctx) error {
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
// @Success      201  {object}  domain.TronWallet
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/tron/wallet [post]
func (h *NewHandler) CreateTronWallet(ctx *fiber.Ctx) error {
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
func (h *NewHandler) CheckTronNetwork(ctx *fiber.Ctx) error {
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
func (h *NewHandler) EstimateTronGas(ctx *fiber.Ctx) error {
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
func (h *NewHandler) GetRPCStatus(ctx *fiber.Ctx) error {
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
func (h *NewHandler) GetAvailableMethods(ctx *fiber.Ctx) error {
	methods := domain.GetAvailableTronRPCMethods()

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
func (h *NewHandler) CallRPCMethod(ctx *fiber.Ctx) error {
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
