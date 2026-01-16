package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
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
			xml_content, sri_status, sri_message, environment, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		er.TransactionID, er.IssuerID, er.TaxPayerID, er.AccessKey, er.ReceiptType,
		er.XMLContent, er.SRIStatus, er.SRIMessage, er.Environment, now, now,
	).Scan(&er.ID, &er.CreatedAt, &er.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create electronic receipt: %w", err)
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

func (r *ElectronicReceiptRepositoryImpl) GetByAccessKey(ctx context.Context, accessKey string) (*domain.ElectronicReceipt, error) {
	query := `
		SELECT id, transaction_id, issuer_id, tax_payer_id, access_key, receipt_type, 
		       xml_content, authorization_date, sri_status, sri_message, ride_path, environment, created_at, updated_at
		FROM electronic_receipts
		WHERE access_key = $1
	`
	var er domain.ElectronicReceipt
	// Manejo de nulos si authorization_date o ride_path son nulos
	var authDate *time.Time
	var ridePath *string

	err := r.db.QueryRow(ctx, query, accessKey).Scan(
		&er.ID, &er.TransactionID, &er.IssuerID, &er.TaxPayerID, &er.AccessKey, &er.ReceiptType,
		&er.XMLContent, &authDate, &er.SRIStatus, &er.SRIMessage, &ridePath, &er.Environment, &er.CreatedAt, &er.UpdatedAt,
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
	// Si un comprobante lleva más de 2 días atascado en PENDIENTE/EN PROCESO, se considera abandonado por el job automático
	// y requerirá intervención manual del usuario (ej: anular y reemitir).
	query := `
		SELECT id, transaction_id, issuer_id, tax_payer_id, access_key, receipt_type, 
		       xml_content, authorization_date, sri_status, sri_message, environment, created_at, updated_at
		FROM electronic_receipts
		WHERE sri_status IN ('PENDIENTE', 'RECIBIDA', 'EN PROCESO', 'ERROR_ENVIO', 'ERROR_RED')
		AND created_at > NOW() - INTERVAL '2 days'
		ORDER BY created_at ASC
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
		// Scan sin ride_path para simplificar, o agregarlo si es necesario
		err := rows.Scan(
			&er.ID, &er.TransactionID, &er.IssuerID, &er.TaxPayerID, &er.AccessKey, &er.ReceiptType,
			&er.XMLContent, &authDate, &er.SRIStatus, &er.SRIMessage, &er.Environment, &er.CreatedAt, &er.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan receipt: %w", err)
		}
		er.AuthorizationDate = authDate
		receipts = append(receipts, er)
	}
	return receipts, nil
}
