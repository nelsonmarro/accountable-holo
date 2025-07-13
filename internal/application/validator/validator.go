// Package validator provides validation logic for the application.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Validator is a struct that holds the target value and a slice of errors.
type Validator struct {
	target any
	errs   []error
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{}
}

// For sets the target value for validation and resets the errors slice.
func (v *Validator) For(obj any) *Validator {
	v.target = obj
	v.errs = []error{}
	return v
}

// Required checks if the specified fields are non-zero/non-empty.
func (v *Validator) Required(fieldNames ...string) *Validator {
	val, err := v.getValidTarget()
	if err != nil {
		v.errs = append(v.errs, err)
		return v
	}

	for _, fieldName := range fieldNames {
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			v.errs = append(v.errs, fmt.Errorf("field '%s' not found in struct", fieldName))
			continue
		}

		isZero := false
		switch field.Kind() {
		case reflect.String:
			isZero = field.Len() == 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isZero = field.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			isZero = field.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isZero = field.Float() == 0.0
		case reflect.Bool:
			isZero = !field.Bool() // Check if false
		case reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
			isZero = field.IsNil()
		default:
			// For other types, check if it's the zero value
			isZero = reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
		}

		if isZero {
			v.errs = append(v.errs, fmt.Errorf("el campo %s es requerido", fieldName))
		}
	}
	return v
}

// NumberMin checks if the specified numeric fields are greater than or equal to the given min value.
func (v *Validator) NumberMin(min float64, fieldNames ...string) *Validator {
	val, err := v.getValidTarget()
	if err != nil {
		v.errs = append(v.errs, err)
		return v
	}

	for _, fieldName := range fieldNames {
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			v.errs = append(v.errs, fmt.Errorf("field '%s' not found in struct", fieldName))
			continue
		}

		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if float64(field.Int()) < min {
				v.errs = append(v.errs, fmt.Errorf("%s debe ser almenos %v", fieldName, min))
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if float64(field.Uint()) < min {
				v.errs = append(v.errs, fmt.Errorf("%s debe ser almenos %v", fieldName, min))
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() < min {
				v.errs = append(v.errs, fmt.Errorf("%s debe ser almenos %v", fieldName, min))
			}
		default:
			v.errs = append(v.errs, fmt.Errorf("el campo '%s' no es de tipo numerico", fieldName))
		}
	}
	return v
}

func (v *Validator) IsDate(fieldName string) *Validator {
	val, err := v.getValidTarget()
	if err != nil {
		v.errs = append(v.errs, err)
		return v
	}
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		v.errs = append(v.errs, fmt.Errorf("field '%s' not found in struct", fieldName))
		return v
	}
	if field.Kind() != reflect.String {
		v.errs = append(v.errs, fmt.Errorf("el campo '%s' debe ser de tipo string para validación de fecha", fieldName))
		return v
	}

	_, err = time.Parse("01/02/2006", field.String())
	if err != nil {
		v.errs = append(v.errs, fmt.Errorf("el campo '%s' no es una fecha válida: %v", fieldName, err))
	}

	return v
}

// Validate returns a slice of errors, or nil if no errors were found.
func (v *Validator) Validate() []error {
	if len(v.errs) == 0 {
		return nil
	}
	return v.errs
}

// ConsolidateErrors returns all errors consolidated into a single string,
func (v *Validator) ConsolidateErrors() error {
	if len(v.errs) == 0 {
		return nil
	}

	errorStrings := make([]string, len(v.errs))
	for i, err := range v.errs {
		errorStrings[i] = err.Error()
	}

	return errors.New(strings.Join(errorStrings, "\n"))
}

func (v *Validator) getValidTarget() (reflect.Value, error) {
	if v.target == nil {
		return reflect.Value{}, fmt.Errorf("validation target not set. Call For() first")
	}

	val := reflect.ValueOf(v.target)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("validation target must be a struct")
	}

	return val, nil
}
