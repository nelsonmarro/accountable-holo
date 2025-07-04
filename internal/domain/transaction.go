package domain

import "time"

type Transaction struct {
	BaseEntity
	Amount          float64   `db:"amount"`
	Description     string    `db:"description"`
	TransactionDate time.Time `db:"transaction_date"`

	// Relationships
	AccountID  int64 `db:"account_id"`
	CategoryID int64 `db:"category_id"`

	Account  *Account
	Category *Category

	// Calculated field
	RunningBalance float64 `db:"-"`
}
