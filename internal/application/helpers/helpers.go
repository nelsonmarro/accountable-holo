// Package helpers provides utility functions for the application.
package helpers

import (
	"regexp"
	"strings"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

func GetAccountTypeFromString(accType string) domain.AccountType {
	switch accType {
	case "Ahorros":
		return domain.SavingAccount
	case "Corriente":
		return domain.OrdinaryAccount
	default:
		return domain.SavingAccount // Default to SavingAcount if unknown
	}
}

func GetCategoryTypeFromString(catType string) domain.CategoryType {
	switch catType {
	case "Ingreso":
		return domain.Income
	case "Egreso":
		return domain.Outcome
	default:
		return domain.Income
	}
}

// A regular expression to find one or more whitespace
// characters (including newlines, tabs, etc.)
var whitespaceRegex = regexp.MustCompile(`\s+`)

// PrepareForTruncation takes a string that may contain new lines and other
// excess whitespace and sanitizes it into a single line suitable for display
func PrepareForTruncation(s string) string {
	// 1. Replace all newline characters (\n, \r\n, etc.) and tabs with a single space.
	// The regex `\s+` matches any sequence of one or more whitespace characters.
	singleLine := whitespaceRegex.ReplaceAllString(s, " ")

	// Trim any leading or trailing whitespace
	return strings.TrimSpace(singleLine)
}

// TruncateString shortens a string to a max length and adds ellipsis.
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}
