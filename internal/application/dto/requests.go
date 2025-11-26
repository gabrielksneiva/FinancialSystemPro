package dto

type UserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type DepositRequest struct {
	Amount      string `json:"amount" validate:"required,numeric,gt=0"`
	CallbackURL string `json:"callback_url" validate:"omitempty,url"`
}

type BalanceRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type WithdrawRequest struct {
	Amount       string `json:"amount" validate:"required,numeric,gt=0"`
	CallbackURL  string `json:"callback_url" validate:"omitempty,url"`
	WithdrawType string `json:"withdraw_type" validate:"omitempty,oneof=internal tron ethereum bitcoin"`
	Chain        string `json:"chain" validate:"omitempty,oneof=tron ethereum bitcoin"` // alternativa explícita quando withdraw_type='tron' ou 'ethereum' ou 'bitcoin'
}

type TransferRequest struct {
	Amount      string `json:"amount" validate:"required,numeric,gt=0"`
	To          string `json:"to" validate:"required,email"`
	CallbackURL string `json:"callback_url" validate:"omitempty,url"`
}

// TronRequest para operações blockchain
type TronRequest struct {
	Address string `json:"address" validate:"required,len=34"`
	Amount  string `json:"amount" validate:"required,numeric,gt=0"`
}

// CreateTronWalletRequest para criar nova carteira
type CreateTronWalletRequest struct {
	// Vazio - gerado automaticamente
}

// GenerateWalletRequest para gerar wallet multi-chain
type GenerateWalletRequest struct {
	Chain string `json:"chain" validate:"required,oneof=tron ethereum"`
}

// SendTronRequest para enviar TRON
type SendTronRequest struct {
	ToAddress string `json:"to_address" validate:"required,len=34"`
	Amount    string `json:"amount" validate:"required,numeric,gt=0"`
}

// EstimateGasRequest para estimar energia TRON
type EstimateGasRequest struct {
	ToAddress string `json:"to_address" validate:"required,len=34"`
	Amount    string `json:"amount" validate:"required,numeric,gt=0"`
}
