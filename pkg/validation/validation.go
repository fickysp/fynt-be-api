package validation

import (
	"github.com/go-playground/validator/v10"
)

type StructValidator struct {
	validate *validator.Validate
}

func NewStructValidator() *StructValidator {
	return &StructValidator{validate: validator.New()}
}

func (v *StructValidator) Validate(out any) error {
	return v.validate.Struct(out)
}
