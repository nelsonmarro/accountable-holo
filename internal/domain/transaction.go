package domain

import "time"

type Transaction struct {
	BaseEntity
	TransactionNumber     string    `db:"transaction_number"`
	Amount                float64   `db:"amount"`
	Description           string    `db:"description"`
	TransactionDate       time.Time `db:"transaction_date"`
	AttachmentPath        *string   `db:"attachment_path"`
	IsVoided              bool      `db:"is_voided"`
	VoidedByTransactionID *int      `db:"voided_by_transaction_id"`
	VoidsTransactionID    *int      `db:"voids_transaction_id"`

	// Relationships
	AccountID  int `db:"account_id"`
	CategoryID int `db:"category_id"`

	Account  *Account
	Category *Category

	// Calculated fields
	RunningBalance      float64 `db:"-"`
	AbsoluteAttachPath string    `db:"-"`
}
