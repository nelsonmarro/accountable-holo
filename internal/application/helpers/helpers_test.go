package helpers

import (
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetDisplayAccountTypeName(t *testing.T) {
	// Define the test cases in a "table" (a slice of structs).
	// Each struct represents one test case.
	testCases := []struct {
		name     string             // A descriptive name for the test case
		input    domain.AccountType // The input to the function
		expected string             // The expected output
	}{
		{
			name:     "Should return 'Ahorros' for SavingAcount",
			input:    domain.SavingAccount,
			expected: "Ahorros",
		},
		{
			name:     "Should return 'Corriente' for OrdinaryAccount",
			input:    domain.OrdinaryAccount,
			expected: "Corriente",
		},
		{
			name:     "Should return 'Cuenta Desconocida' for an unknown type",
			input:    domain.AccountType("some_other_type"), // Test the default case
			expected: "Cuenta Desconocida",
		},
		{
			name:     "Should return 'Cuenta Desconocida' for an empty type",
			input:    "",
			expected: "Cuenta Desconocida",
		},
	}

	// Loop through all the test cases.
	for _, tc := range testCases {
		// t.Run() creates a sub-test for each case, which gives clearer test output.
		t.Run(tc.name, func(t *testing.T) {
			// Act: Call the function we are testing.
			actual := GetDisplayAccountTypeName(tc.input)

			// Assert: Check if the actual output matches the expected output.
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetAccountTypeFromString(t *testing.T) {
	// Define the table for the second function.
	testCases := []struct {
		name     string
		input    string
		expected domain.AccountType
	}{
		{
			name:     "Should return SavingAcount for 'Ahorros'",
			input:    "Ahorros",
			expected: domain.SavingAccount,
		},
		{
			name:     "Should return OrdinaryAccount for 'Corriente'",
			input:    "Corriente",
			expected: domain.OrdinaryAccount,
		},
		{
			name:     "Should default to SavingAcount for an unknown string",
			input:    "some_random_string",
			expected: domain.SavingAccount, // Verify the default case
		},
		{
			name:     "Should default to SavingAcount for an empty string",
			input:    "",
			expected: domain.SavingAccount, // Verify the default case
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			actual := GetAccountTypeFromString(tc.input)

			// Assert
			assert.Equal(t, tc.expected, actual)
		})
	}
}
