package domain

type TaxPayer struct {
	BaseEntity
	Identification     string `db:"identification"`
	IdentificationType string `db:"identification_type"`
	Name               string `db:"name"`
	Email              string `db:"email"`
	Address            string `db:"address"`
	Phone              string `db:"phone"`
}
