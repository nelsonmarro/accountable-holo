package domain

// TransactionType defines if a transaction is an income or an outcome.
type TransactionType string

const (
	Income  TransactionType = "Ingreso"
	Outcome TransactionType = "Egreso"
)

type Category struct {
	BaseEntity
	Type TransactionType `db:"type"`
	Name string          `db:"name"`
}
