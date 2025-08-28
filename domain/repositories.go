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
	Amount string `json:"amount" validate:"required"`
}

type BalanceRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

type WithdrawRequest struct {
	Amount string `json:"amount" validate:"required"`
}
