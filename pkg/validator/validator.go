package validator

import (
	"life_easy_api/pkg/db-pkg/models"

	playground "github.com/go-playground/validator/v10"
)

var instance = playground.New(
	playground.WithRequiredStructEnabled(),
)

func Check(input any) error {
	if err := instance.Struct(input); err != nil {
		return models.ErrInvalid
	}
	return nil
}
