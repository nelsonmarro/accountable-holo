package domain

import "time"

type ElectronicReceipt struct {
	BaseEntity
	TransactionID     int        `db:"transaction_id"`
	IssuerID          int        `db:"issuer_id"`
	TaxPayerID        int        `db:"tax_payer_id"`
	AccessKey         string     `db:"access_key"`
	ReceiptType       string     `db:"receipt_type"`
	XMLContent        string     `db:"xml_content"`
	AuthorizationDate *time.Time `db:"authorization_date"`
	SRIStatus         string     `db:"sri_status"`
	SRIMessage        string     `db:"sri_message"`
	RidePath          string     `db:"ride_path"`
	Environment       int        `db:"environment"`

	// Relaciones (opcionales para carga en memoria)
	Issuer   *Issuer   `db:"-"`
	TaxPayer *TaxPayer `db:"-"`
}
