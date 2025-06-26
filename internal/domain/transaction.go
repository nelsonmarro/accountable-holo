package domain

import "time"

// TransactionType defines if a transaction is an income or an outcome.
type TransactionType string

const (
	Income  TransactionType = "income"
	Outcome TransactionType = "outcome"
)

type Transaction struct {
	BaseEntity
	Amount          float64         `db:"amount"`
	Type            TransactionType `db:"type"`
	Description     string          `db:"description"`
	TransactionDate time.Time       `db:"transaction_date"`

	// Relationships
	AccountID  int64 `db:"account_id"`
	CategoryID int64 `db:"category_id"`

	Account  *Account
	Category *Category
}
