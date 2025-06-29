package uivalidators

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// We'll create one main test function and use sub-tests for each scenario.
func TestUIValidator(t *testing.T) {
	// --- Test Individual Rules ---

	t.Run("Required validator", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       string
			expectError bool
		}{
			{"should pass for non-empty string", "hello", false},
			{"should pass for string with spaces", "  hello  ", false},
			{"should fail for empty string", "", true},
			{"should fail for string with only whitespace", "   ", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				v := NewValidator()
				v.Required() // Add the rule to be tested

				// Act
				err := v.Validate(tc.input)

				// Assert
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("IsInt validator", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       string
			expectError bool
		}{
			{"should pass for positive integer", "123", false},
			{"should pass for negative integer", "-45", false},
			{"should pass for zero", "0", false},
			{"should fail for non-integer string", "abc", true},
			{"should fail for float string", "123.45", true},
			{"should fail for empty string", "", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				v := NewValidator()
				v.IsInt()
				err := v.Validate(tc.input)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("MinLength validator", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       string
			minLength   int
			expectError bool
		}{
			{"should pass when length is greater than min", "hello", 4, false},
			{"should pass when length is equal to min", "world", 5, false},
			{"should fail when length is less than min", "hi", 3, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				v := NewValidator()
				v.MinLength(tc.minLength)
				err := v.Validate(tc.input)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	// --- Test Chained Rules ---
	// This is the most important part, testing the builder pattern.

	t.Run("Chained validators", func(t *testing.T) {
		testCases := []struct {
			name        string
			setup       func() *UIValidator // A function to setup the validator chain
			input       string
			expectedErr error // We can check for specific errors
		}{
			{
				name: "Required and IsInt should pass for valid integer string",
				setup: func() *UIValidator {
					v := NewValidator()
					v.Required()
					v.IsInt()
					return v
				},
				input:       "12345",
				expectedErr: nil,
			},
			{
				name: "Required and IsInt should fail on Required for empty string",
				setup: func() *UIValidator {
					v := NewValidator()
					v.Required()
					v.IsInt()
					return v
				},
				input:       "",
				expectedErr: errors.New("el campo no puede estar vacío"),
			},
			{
				name: "Required and IsInt should fail on IsInt for non-integer string",
				setup: func() *UIValidator {
					v := NewValidator()
					v.Required()
					v.IsInt()
					return v
				},
				input:       "not-a-number",
				expectedErr: errors.New("el campo debe ser un número entero válido"),
			},
			{
				name: "IsInt and MinLength(5) should pass for long integer",
				setup: func() *UIValidator {
					v := NewValidator()
					v.IsInt()
					v.MinLength(5)
					return v
				},
				input:       "55555",
				expectedErr: nil,
			},
			{
				name: "IsInt and MinLength(5) should fail on MinLength for short integer",
				setup: func() *UIValidator {
					v := NewValidator()
					v.IsInt()
					v.MinLength(5)
					return v
				},
				input:       "123",
				expectedErr: errors.New("el campo debe tener al menos 5 caracteres"),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				validator := tc.setup()

				// Act
				err := validator.Validate(tc.input)

				// Assert
				// We use require here because if the error doesn't match,
				// there's no point in continuing the test.
				if tc.expectedErr != nil {
					require.Error(t, err)
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

