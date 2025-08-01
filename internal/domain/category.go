package domain

// CategoryType defines if a transaction is an income or an outcome.
type CategoryType string

const (
	Income     CategoryType = "Ingreso"
	Outcome    CategoryType = "Egreso"
	Adjustment CategoryType = "Ajuste"
)

type Category struct {
	BaseEntity
	Type CategoryType `db:"type"`
	Name string       `db:"name"`
}
