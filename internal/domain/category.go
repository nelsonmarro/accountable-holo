package domain

import "github.com/shopspring/decimal"

// CategoryType defines if a transaction is an income or an outcome.
type CategoryType string

const (
	Income  CategoryType = "Ingreso"
	Outcome CategoryType = "Egreso"
)

type Category struct {
	BaseEntity
	Type          CategoryType     `db:"type"`
	Name          string           `db:"name"`
	MonthlyBudget *decimal.Decimal `db:"monthly_budget"`
}
