package helpers

import (
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
)

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
