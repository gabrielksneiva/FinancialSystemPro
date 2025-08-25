package domain
import "github.com/shopspring/decimal"

type UserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type DepositRequest struct{
	Amount decimal.Decimal `json:"amount" validate:"required, gt=0"`
}
