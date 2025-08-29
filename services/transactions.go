package services

import (
	r "financial-system-pro/repositories"
	w "financial-system-pro/workers"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type NewTransactionService struct {
	DB *r.NewDatabase
	W  *w.TransactionWorkerPool
}

func (t *NewTransactionService) Deposit(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	id := c.Locals("ID").(string)

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if t.W != nil {
		job := w.TransactionJob{
			Type:        w.JobDeposit,
			Account:     uid,
			Amount:      amount,
			CallbackURL: callbackURL,
			JobID:       uuid.New(),
		}

		t.W.Jobs <- job

		return &ServiceResponse{
			StatusCode: fiber.StatusAccepted,
			Body: fiber.Map{
				"job_id": job.JobID.String(),
				"status": "pending",
			},
		}, nil
	}

	if err := t.DB.Transaction(uid, amount, "deposit"); err != nil {
		return nil, err
	}
	if err := t.DB.Insert(&r.Transaction{
		AccountID:   uid,
		Amount:      amount,
		Type:        "deposit",
		Category:    "credit",
		Description: "User deposit",
	}); err != nil {
		return nil, err
	}

	return &ServiceResponse{
		StatusCode: fiber.StatusOK,
		Body:       fiber.Map{"message": "Deposit succesfully"},
	}, nil
}

func (t *NewTransactionService) GetBalance(c *fiber.Ctx, userID string) (decimal.Decimal, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return decimal.Zero, err
	}

	response, err := t.DB.Balance(uid)
	if err != nil {
		return decimal.Zero, err
	}

	return response, nil
}

func (t *NewTransactionService) Withdraw(c *fiber.Ctx, amount decimal.Decimal, callbackURL string) (*ServiceResponse, error) {
	id := c.Locals("ID").(string)

	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if t.W != nil {
		job := w.TransactionJob{
			Type:        w.JobWithdraw,
			Account:     uuid,
			Amount:      amount,
			CallbackURL: callbackURL,
		}

		t.W.Jobs <- job
		return &ServiceResponse{
			StatusCode: fiber.StatusAccepted,
			Body: fiber.Map{
				"job_id": job.JobID.String(),
				"status": "pending",
			},
		}, nil
	}

	err = t.DB.Transaction(uuid, amount, "withdraw")
	if err != nil {
		return nil, err
	}

	err = t.DB.Insert(&r.Transaction{
		AccountID:   uuid,
		Amount:      amount,
		Type:        "withdraw",
		Category:    "debit",
		Description: "User withdraw",
	})
	if err != nil {
		return nil, err
	}

	return &ServiceResponse{
		StatusCode: fiber.StatusOK,
		Body:       fiber.Map{"message": "Withdraw succesfully"},
	}, nil
}

func (t *NewTransactionService) Transfer(c *fiber.Ctx, amount decimal.Decimal, userTo, callbackURL string) (*ServiceResponse, error) {
	id := c.Locals("ID").(string)
	userFrom, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if t.W != nil {
		job := w.TransactionJob{
			Type:        w.JobTransfer,
			Account:     userFrom,
			Amount:      amount,
			ToEmail:     userTo,
			CallbackURL: callbackURL,
			JobID:       uuid.New(),
		}

		t.W.Jobs <- job

		return &ServiceResponse{
			StatusCode: fiber.StatusAccepted,
			Body: fiber.Map{
				"job_id": job.JobID.String(),
				"status": "pending",
			},
		}, nil
	}

	// fallback synchronous processing if worker pool is not initialized
	foundUser, err := t.DB.FindUserByField("email", userTo)
	if err != nil {
		return nil, err
	}
	destinyUserID := foundUser.ID

	foundUserFrom, err := t.DB.FindUserByField("id", userFrom.String())
	if err != nil {
		return nil, err
	}

	if err := t.DB.Transaction(userFrom, amount, "withdraw"); err != nil {
		return nil, err
	}
	if err := t.DB.Transaction(destinyUserID, amount, "deposit"); err != nil {
		return nil, err
	}

	if err := t.DB.Insert(&r.Transaction{
		AccountID:   userFrom,
		Amount:      amount,
		Type:        "transfer",
		Category:    "debit",
		Description: "User transfer to " + userTo,
	}); err != nil {
		return nil, err
	}

	if err := t.DB.Insert(&r.Transaction{
		AccountID:   destinyUserID,
		Amount:      amount,
		Type:        "transfer",
		Category:    "credit",
		Description: "User transfer from " + foundUserFrom.Email,
	}); err != nil {
		return nil, err
	}

	return &ServiceResponse{
		StatusCode: fiber.StatusOK,
		Body:       fiber.Map{"message": "Transfer succesfully"},
	}, nil
}
