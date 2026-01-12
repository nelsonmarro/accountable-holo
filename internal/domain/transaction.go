package domain

import "time"

type Transaction struct {
	BaseEntity
	TransactionNumber string    `db:"transaction_number"`
	Amount            float64   `db:"amount"` // Este es el TOTAL (con impuestos)
	TransactionDate   time.Time `db:"transaction_date"`
	Description       string    `db:"description"`
	AccountID         int       `db:"account_id"`
	CategoryID        int       `db:"category_id"`

	// Campos para Facturación Electrónica (SRI)
	Subtotal15 float64 `db:"subtotal_15"`
	Subtotal0  float64 `db:"subtotal_0"`
	TaxAmount  float64 `db:"tax_amount"`
	TaxPayerID *int    `db:"tax_payer_id"` // Puntero para soportar NULL

	// Otros campos existentes...
	AttachmentPath        *string `db:"attachment_path"`
	AbsoluteAttachPath    string  `db:"-"`
	IsVoided              bool    `db:"is_active"`
	VoidedByTransactionID *int    `db:"voided_by_transaction_id"`
	VoidsTransactionID    *int    `db:"voids_transaction_id"`

	// Relaciones
	Category      *Category `db:"-"`
	CreatedByUser *User     `db:"-"`
	UpdatedByUser *User     `db:"-"`
	CreatedByID   int       `db:"created_by_id"`
	UpdatedByID   int       `db:"updated_by_id"`

	RunningBalance float64           `db:"running_balance"`
	Items          []TransactionItem `db:"-"` // Detalle de la transacción

	// Relación con SRI
	ElectronicReceipt *ElectronicReceipt `db:"-"`
}

type TransactionItem struct {
	BaseEntity
	TransactionID int     `db:"transaction_id"`
	Description   string  `db:"description"`
	Quantity      float64 `db:"quantity"`
	UnitPrice     float64 `db:"unit_price"`
	TaxRate       int     `db:"tax_rate"` // 0, 2, 4
	Subtotal      float64 `db:"subtotal"` // unit_price * quantity
}
