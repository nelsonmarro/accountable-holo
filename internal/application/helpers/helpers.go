// Package helpers provides utility functions for the application.
package helpers

import "github.com/nelsonmarro/accountable-holo/internal/domain"

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
