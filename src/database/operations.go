package database

import (
	"github.com/go-playground/validator/v10"
)

type Model interface {
	Validate() error
	Update() error
	Create() error
}

func Insert(m Model) error {
	err := m.Validate()
	if err != nil {
		return err
	}
	err = m.Create()
	if err != nil {
		return err
	}
	return nil
}

func Update(m Model) error {
	err := m.Validate()
	if err != nil {
		return err
	}
	err = m.Update()
	if err != nil {
		return err
	}
	return nil
}

// Validator tags resources: https://godoc.org/github.com/go-playground/validator#hdr-Length

func BuildValidationErrorMsg(errs validator.ValidationErrors) string {
	msg := ""

	for _, err := range errs {
		msg += "Validation failed on field: " + err.Field() + ". Expected: " + err.Tag() + " " + err.Param() + ". Got: " + err.Value().(string) + ".\n"
	}
	return msg
}