package services

import (
	"financial-system-pro/domain"
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

	err = t.Database.Transaction(uid, amount, "deposit")
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

func (t *NewTransactionService) GetBalance(c *fiber.Ctx, userID string) (decimal.Decimal, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return decimal.Zero, err
	}

	response, err := t.Database.Balance(uid)
	if err != nil {
		return decimal.Zero, err
	}

	return response, nil
}

func (t *NewTransactionService) Withdraw(c *fiber.Ctx, amount decimal.Decimal) error {
	id := c.Locals("ID").(string)

	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	err = t.Database.Transaction(uuid, amount, "withdraw")
	if err != nil {
		return err
	}

	err = t.Database.Insert(&repositories.Transaction{
		AccountID:   uuid,
		Amount:      amount,
		Type:        "withdraw",
		Category:    "debit",
		Description: "User withdraw",
	})
	if err != nil {
		return err
	}

	return nil
}

func (t *NewTransactionService) Transfer(c *fiber.Ctx, transferRequest *domain.TransferRequest) error {
	id := c.Locals("ID").(string)

	userFrom, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	amount, err := decimal.NewFromString(transferRequest.Amount)
	if err != nil {
		return err
	}

	foundUser, err := t.Database.FindUserByField("email", transferRequest.To)
	if err != nil {
		return err
	}

	userTo := foundUser.ID

	err = t.Database.Transaction(userFrom, amount, "withdraw")
	if err != nil {
		return err
	}

	err = t.Database.Transaction(userFrom, amount, "deposit")
	if err != nil {
		return err
	}

	err = t.Database.Insert(&repositories.Transaction{
		AccountID:   userFrom,
		Amount:      amount,
		Type:        "transfer",
		Category:    "debit",
		Description: "User transfer to " + transferRequest.To,
	})
	if err != nil {
		return err
	}

	err = t.Database.Insert(&repositories.Transaction{
		AccountID:   userTo,
		Amount:      amount,
		Type:        "transfer",
		Category:    "credit",
		Description: "User transfer from" + id,
	})
	if err != nil {
		return err
	}

	return nil
}
