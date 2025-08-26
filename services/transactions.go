package services

import (
	"financial-system-pro/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type NewTransactionService struct {
	Database *repositories.NewDatabase
}

func (t *NewTransactionService) Deposit(c *fiber.Ctx, amount decimal.Decimal) error {
	id := c.Locals("ID").(string)

	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	err = t.Database.Deposit(uid, amount)
	if err != nil {
		return err
	}

	err = t.Database.Insert(&repositories.Transaction{
		AccountID:   uid,
		Amount:      amount,
		Type:        "deposit",
		Category:    "credit",
		Description: "User deposit",
	})
	if err != nil {
		return err
	}

	return nil
}
