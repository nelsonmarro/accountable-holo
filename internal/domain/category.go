package domain

// TransactionType defines if a transaction is an income or an outcome.
type TransactionType string

const (
	Income  TransactionType = "income"
	Outcome TransactionType = "outcome"
)

type Category struct {
	BaseEntity
	Type TransactionType `db:"type"`
	Name string          `db:"name"`
}
