// Package domain.
package domain

type AccountType string

const (
	SavingAccount   AccountType = "Ahorros"
	OrdinaryAccount AccountType = "Corriente"
)

type Account struct {
	BaseEntity
	Name           string      `db:"name"`
	Number         string      `db:"number"`
	Type           AccountType `db:"type"`
	InitialBalance float64     `db:"initial_balance"`
}
