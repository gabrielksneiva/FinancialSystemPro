package services

import (
	"financial-system-pro/domain"
	"financial-system-pro/repositories"

	"github.com/gofiber/fiber/v2"
)

type NewTransactionService struct {
	Database *repositories.NewDatabase
}

func (t *NewTransactionService) Deposit(c *fiber.Ctx, depositData *domain.DepositRequest) error {
	id := c.Locals("ID").(string)

	err := t.Database.Deposit(id, depositData.Amount)
	if err != nil {
		return err
	}

	return nil
}