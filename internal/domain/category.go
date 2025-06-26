package domain

type Category struct {
	BaseEntity
	Name string `db:"name"`
}
