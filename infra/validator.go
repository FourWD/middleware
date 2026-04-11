package infra

import (
	"errors"
	"fmt"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *govalidator.Validate
}

func NewValidator() *Validator {
	return &Validator{validate: govalidator.New()}
}

// Validate implements fiber.StructValidator so this can be passed as
// fiber.Config.StructValidator to activate validate tags on c.Bind().Body().
func (v *Validator) Validate(value any) error {
	return v.Struct(value)
}

func (v *Validator) Struct(value any) error {
	if err := v.validate.Struct(value); err != nil {
		validationErrors, ok := err.(govalidator.ValidationErrors)
		if !ok {
			return err
		}

		parts := make([]string, 0, len(validationErrors))
		for _, validationError := range validationErrors {
			parts = append(parts, fmt.Sprintf("%s failed on %s", validationError.Field(), validationError.Tag()))
		}

		return errors.New(strings.Join(parts, ", "))
	}

	return nil
}
