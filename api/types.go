package api

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func isValid(ctx *fiber.Ctx, body interface{}) error {
	err := ctx.BodyParser(body)
	if err != nil {
		return fmt.Errorf("invalid JSON: %s", err)
	}

	return validate.Struct(body)
}
