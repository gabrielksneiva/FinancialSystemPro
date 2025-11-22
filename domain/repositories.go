package domain

type UserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type DepositRequest struct {
	Amount      string `json:"amount" validate:"required"`
	CallbackURL string `json:"callback_url" validate:"omitempty,url"`
}

type BalanceRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

type WithdrawRequest struct {
	Amount       string `json:"amount" validate:"required"`
	CallbackURL  string `json:"callback_url" validate:"omitempty,url"`
	TronAddress  string `json:"tron_address" validate:"omitempty,required_if=WithdrawType tron"` // Para withdraw TRON
	WithdrawType string `json:"withdraw_type" validate:"omitempty,oneof=internal tron"`          // "internal" ou "tron"
}

type TransferRequest struct {
	Amount      string `json:"amount" validate:"required"`
	To          string `json:"to" validate:"required,email"`
	CallbackURL string `json:"callback_url" validate:"omitempty,url"`
}
