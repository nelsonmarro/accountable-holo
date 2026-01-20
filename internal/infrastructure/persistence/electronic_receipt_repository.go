package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/verith/internal/domain"
)

type ElectronicReceiptRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewElectronicReceiptRepository(db *pgxpool.Pool) *ElectronicReceiptRepositoryImpl {
	return &ElectronicReceiptRepositoryImpl{db: db}
}

func (r *ElectronicReceiptRepositoryImpl) Create(ctx context.Context, er *domain.ElectronicReceipt) error {
	query := `
		INSERT INTO electronic_receipts (
			transaction_id, issuer_id, tax_payer_id, access_key, receipt_type, 
			xml_content, sri_status, sri_message, environment, email_sent, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		er.TransactionID, er.IssuerID, er.TaxPayerID, er.AccessKey, er.ReceiptType,
		er.XMLContent, er.SRIStatus, er.SRIMessage, er.Environment, er.EmailSent, now, now,
	).Scan(&er.ID, &er.CreatedAt, &er.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create electronic receipt: %w", err)
	}
	return nil
}

func (r *ElectronicReceiptRepositoryImpl) Update(ctx context.Context, er *domain.ElectronicReceipt) error {
	query := `
		UPDATE electronic_receipts SET 
			access_key = $1, xml_content = $2, sri_status = $3, sri_message = $4, 
			environment = $5, authorization_date = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.db.Exec(ctx, query,
		er.AccessKey, er.XMLContent, er.SRIStatus, er.SRIMessage,
		er.Environment, er.AuthorizationDate, time.Now(), er.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update electronic receipt: %w", err)
	}
	return nil
}

func (r *ElectronicReceiptRepositoryImpl) UpdateStatus(ctx context.Context, accessKey string, status string, message string, authDate *time.Time) error {
	query := `
		UPDATE electronic_receipts SET 
			sri_status = $1, sri_message = $2, authorization_date = $3, updated_at = $4
		WHERE access_key = $5
	`
	_, err := r.db.Exec(ctx, query, status, message, authDate, time.Now(), accessKey)
	if err != nil {
		return fmt.Errorf("failed to update receipt status: %w", err)
	}
	return nil
}

func (r *ElectronicReceiptRepositoryImpl) UpdateXML(ctx context.Context, accessKey string, xmlContent string) error {
	query := `UPDATE electronic_receipts SET xml_content = $1, updated_at = $2 WHERE access_key = $3`
	_, err := r.db.Exec(ctx, query, xmlContent, time.Now(), accessKey)
	return err
}

func (r *ElectronicReceiptRepositoryImpl) UpdateTaxPayerID(ctx context.Context, accessKey string, taxPayerID int) error {
	query := `UPDATE electronic_receipts SET tax_payer_id = $1, updated_at = $2 WHERE access_key = $3`
	_, err := r.db.Exec(ctx, query, taxPayerID, time.Now(), accessKey)
	return err
}

func (r *ElectronicReceiptRepositoryImpl) UpdateEmailSent(ctx context.Context, accessKey string, sent bool) error {
	query := `UPDATE electronic_receipts SET email_sent = $1, updated_at = $2 WHERE access_key = $3`
	_, err := r.db.Exec(ctx, query, sent, time.Now(), accessKey)
	return err
}

func (r *ElectronicReceiptRepositoryImpl) GetByAccessKey(ctx context.Context, accessKey string) (*domain.ElectronicReceipt, error) {
	query := `
		SELECT id, transaction_id, issuer_id, tax_payer_id, access_key, receipt_type, 
		       xml_content, authorization_date, sri_status, sri_message, ride_path, environment, email_sent, created_at, updated_at
		FROM electronic_receipts
		WHERE access_key = $1
	`
	var er domain.ElectronicReceipt
	// Manejo de nulos si authorization_date o ride_path son nulos
	var authDate *time.Time
	var ridePath *string

	err := r.db.QueryRow(ctx, query, accessKey).Scan(
		&er.ID, &er.TransactionID, &er.IssuerID, &er.TaxPayerID, &er.AccessKey, &er.ReceiptType,
		&er.XMLContent, &authDate, &er.SRIStatus, &er.SRIMessage, &ridePath, &er.Environment, &er.EmailSent, &er.CreatedAt, &er.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get receipt by access key: %w", err)
	}

	er.AuthorizationDate = authDate
	if ridePath != nil {
		er.RidePath = *ridePath
	}

	return &er, nil
}

func (r *ElectronicReceiptRepositoryImpl) FindPendingReceipts(ctx context.Context) ([]domain.ElectronicReceipt, error) {
	// Solo intentamos sincronizar comprobantes recientes (últimos 2 días).
	// Hacemos JOIN para mostrar datos útiles al usuario (Cliente, Monto, Nro Factura)
	query := `
		SELECT r.id, r.transaction_id, r.issuer_id, r.tax_payer_id, r.access_key, r.receipt_type, 
		       r.xml_content, r.authorization_date, r.sri_status, r.sri_message, r.environment, r.email_sent, r.created_at, r.updated_at,
			   t.transaction_number, t.amount, COALESCE(tp.name, 'CONSUMIDOR FINAL')
		FROM electronic_receipts r
		JOIN transactions t ON r.transaction_id = t.id
		LEFT JOIN tax_payers tp ON r.tax_payer_id = tp.id
		WHERE r.sri_status IN ('PENDIENTE', 'RECIBIDA', 'EN PROCESO', 'ERROR_ENVIO', 'ERROR_RED')
		AND r.created_at > NOW() - INTERVAL '2 days'
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending receipts: %w", err)
	}
	defer rows.Close()

	var receipts []domain.ElectronicReceipt
	for rows.Next() {
		var er domain.ElectronicReceipt
		var authDate *time.Time
		
		err := rows.Scan(
			&er.ID, &er.TransactionID, &er.IssuerID, &er.TaxPayerID, &er.AccessKey, &er.ReceiptType,
			&er.XMLContent, &authDate, &er.SRIStatus, &er.SRIMessage, &er.Environment, &er.EmailSent, &er.CreatedAt, &er.UpdatedAt,
			&er.TransactionNumber, &er.TotalAmount, &er.ClientName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan receipt: %w", err)
		}
		er.AuthorizationDate = authDate
		receipts = append(receipts, er)
	}
	return receipts, nil
}
