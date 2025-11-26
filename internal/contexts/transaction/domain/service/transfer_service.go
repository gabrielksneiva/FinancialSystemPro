package service

import (
	"context"
	"errors"
	"financial-system-pro/internal/contexts/transaction/domain/entity"
	userentity "financial-system-pro/internal/contexts/user/domain/entity"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransferService coordena transferências entre wallets de usuários
// Domain Service para lógica cross-aggregate
type TransferService struct {
	// Dependências serão injetadas
}

// NewTransferService cria uma nova instância do serviço
func NewTransferService() *TransferService {
	return &TransferService{}
}

// TransferRequest representa uma solicitação de transferência
type TransferRequest struct {
	FromUserID uuid.UUID
	ToUserID   uuid.UUID
	Amount     decimal.Decimal
	Reference  string
}

// TransferResult representa o resultado de uma transferência
type TransferResult struct {
	TransactionID uuid.UUID
	FromBalance   decimal.Decimal
	ToBalance     decimal.Decimal
	Success       bool
	ErrorMessage  string
}

// ValidateTransfer valida uma solicitação de transferência
func (s *TransferService) ValidateTransfer(ctx context.Context, request TransferRequest) error {
	if request.FromUserID == uuid.Nil {
		return errors.New("source user ID cannot be empty")
	}

	if request.ToUserID == uuid.Nil {
		return errors.New("destination user ID cannot be empty")
	}

	if request.FromUserID == request.ToUserID {
		return errors.New("cannot transfer to the same user")
	}

	if request.Amount.IsNegative() || request.Amount.IsZero() {
		return errors.New("transfer amount must be positive")
	}

	return nil
}

// ExecuteTransfer executa uma transferência entre dois agregados de usuário
func (s *TransferService) ExecuteTransfer(
	ctx context.Context,
	fromAggregate *userentity.UserAggregate,
	toAggregate *userentity.UserAggregate,
	amount decimal.Decimal,
	reference string,
) (*entity.TransactionAggregate, error) {

	// Validações de negócio
	if !fromAggregate.CanWithdraw(amount.InexactFloat64()) {
		return nil, errors.New("insufficient funds or inactive source user")
	}

	if !toAggregate.User().IsActive() {
		return nil, errors.New("destination user is inactive")
	}

	// Debita do usuário origem
	if err := fromAggregate.DebitWallet(amount.InexactFloat64()); err != nil {
		return nil, err
	}

	// Credita no usuário destino
	if err := toAggregate.CreditWallet(amount.InexactFloat64()); err != nil {
		// Rollback: recredita o valor debitado
		_ = fromAggregate.CreditWallet(amount.InexactFloat64())
		return nil, err
	}

	// Cria agregado de transação para registro
	txAgg, err := entity.NewTransactionAggregate(
		fromAggregate.User().ID,
		entity.TransactionTypeTransfer,
		amount,
		"USD",
		"internal",
		toAggregate.User().ID.String(),
	)
	if err != nil {
		// Rollback completo
		_ = fromAggregate.CreditWallet(amount.InexactFloat64())
		_ = toAggregate.DebitWallet(amount.InexactFloat64())
		return nil, err
	}

	// Marca como confirmada e completa imediatamente (transferência interna)
	_ = txAgg.Confirm("internal-transfer", 1, 0)
	_ = txAgg.Complete(amount)

	return txAgg, nil
}

// CalculateFee calcula a taxa de transferência
func (s *TransferService) CalculateFee(amount decimal.Decimal, transferType string) decimal.Decimal {
	// Transferências internas não têm taxa
	if transferType == "internal" {
		return decimal.Zero
	}

	// Taxa de 1% para outras transferências
	feePercentage := decimal.NewFromFloat(0.01)
	return amount.Mul(feePercentage)
}

// ValidateBalance verifica se o usuário tem saldo suficiente incluindo taxa
func (s *TransferService) ValidateBalance(
	userAggregate *userentity.UserAggregate,
	amount decimal.Decimal,
	transferType string,
) error {
	fee := s.CalculateFee(amount, transferType)
	totalRequired := amount.Add(fee)

	if !userAggregate.Wallet().HasSufficientBalance(totalRequired.InexactFloat64()) {
		return errors.New("insufficient balance including fees")
	}

	return nil
}
