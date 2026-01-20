package domain

type EmissionPoint struct {
	BaseEntity
	IssuerID          int    `db:"issuer_id"`
	EstablishmentCode string `db:"establishment_code"`
	EmissionPointCode string `db:"emission_point_code"`
	ReceiptType       string `db:"receipt_type"`
	CurrentSequence   int    `db:"current_sequence"`
	InitialSequence   int    `db:"initial_sequence"`
	IsActive          bool   `db:"is_active"`
}
