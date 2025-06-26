// Package domain.
package domain

type AccountType string

const (
	SavingAcount    AccountType = "bank_saving"
	OrdinaryAccount AccountType = "bank_ordinary"
)

type Account struct {
	BaseEntity
	Name           string      `db:"name"`
	Number         string      `db:"number"`
	Type           AccountType `db:"type"`
	InitialBalance float64     `db:"initial_balance"`
}
