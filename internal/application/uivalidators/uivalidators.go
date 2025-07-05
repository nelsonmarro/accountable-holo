// Package uivalidators containes validators helper methods.
package uivalidators

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type UIValidator struct {
	validatorFuncs []func(s string) error
}

func NewValidator() *UIValidator {
	return &UIValidator{}
}

func (v *UIValidator) Required() {
	validatorFunc := func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("el campo no puede estar vacío")
		}
		return nil
	}
	v.validatorFuncs = append(v.validatorFuncs, validatorFunc)
}

func (v *UIValidator) IsInt() {
	validatorFunc := func(s string) error {
		_, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("el campo debe ser un número entero válido")
		}
		return nil
	}
	v.validatorFuncs = append(v.validatorFuncs, validatorFunc)
}

func (v *UIValidator) IsFloat() {
	validatorFunc := func(s string) error {
		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New("el campo debe ser un número decimal válido")
		}
		return nil
	}
	v.validatorFuncs = append(v.validatorFuncs, validatorFunc)
}

func (v *UIValidator) MinLength(min int) {
	validatorFunc := func(s string) error {
		if len(s) < min {
			return errors.New("el campo debe tener al menos " + strconv.Itoa(min) + " caracteres")
		}
		return nil
	}
	v.validatorFuncs = append(v.validatorFuncs, validatorFunc)
}

func (v *UIValidator) IsDate() {
	validatorFunc := func(s string) error {
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return errors.New("el campo debe ser una fecha válida en formato YYYY-MM-DD")
		}
		return nil
	}
	v.validatorFuncs = append(v.validatorFuncs, validatorFunc)
}

func (v *UIValidator) Validate(s string) error {
	for _, fn := range v.validatorFuncs {
		if err := fn(s); err != nil {
			// If it returns an error, stop immediately and return that error.
			return err
		}
	}
	return nil
}
