package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TaxPayerRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewTaxPayerRepository(db *pgxpool.Pool) *TaxPayerRepositoryImpl {
	return &TaxPayerRepositoryImpl{db: db}
}

func (r *TaxPayerRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.TaxPayer, error) {
	query := `
		SELECT id, identification, identification_type, name, email, COALESCE(address, ''), COALESCE(phone, ''), created_at, updated_at
		FROM tax_payers
		WHERE id = $1
	`
	var tp domain.TaxPayer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tp.ID, &tp.Identification, &tp.IdentificationType, &tp.Name, &tp.Email, &tp.Address, &tp.Phone, &tp.CreatedAt, &tp.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get taxpayer by ID: %w", err)
	}
	return &tp, nil
}

func (r *TaxPayerRepositoryImpl) GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error) {
	query := `
		SELECT id, identification, identification_type, name, email, COALESCE(address, ''), COALESCE(phone, ''), created_at, updated_at
		FROM tax_payers
		WHERE identification = $1
	`
	var tp domain.TaxPayer
	err := r.db.QueryRow(ctx, query, identification).Scan(
		&tp.ID, &tp.Identification, &tp.IdentificationType, &tp.Name, &tp.Email, &tp.Address, &tp.Phone, &tp.CreatedAt, &tp.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get taxpayer by identification: %w", err)
	}
	return &tp, nil
}

func (r *TaxPayerRepositoryImpl) Create(ctx context.Context, tp *domain.TaxPayer) error {
	query := `
		INSERT INTO tax_payers (identification, identification_type, name, email, address, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		tp.Identification, tp.IdentificationType, tp.Name, tp.Email, tp.Address, tp.Phone, now, now,
	).Scan(&tp.ID, &tp.CreatedAt, &tp.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create taxpayer: %w", err)
	}
	return nil
}

func (r *TaxPayerRepositoryImpl) Update(ctx context.Context, tp *domain.TaxPayer) error {
	query := `
		UPDATE tax_payers SET
			name=$1, email=$2, address=$3, phone=$4, identification_type=$5, updated_at=$6
		WHERE id=$7
	`
	now := time.Now()
	_, err := r.db.Exec(ctx, query, tp.Name, tp.Email, tp.Address, tp.Phone, tp.IdentificationType, now, tp.ID)
	if err != nil {
		return fmt.Errorf("failed to update taxpayer: %w", err)
	}
	tp.UpdatedAt = now
	return nil
}

func (r *TaxPayerRepositoryImpl) GetAll(ctx context.Context) ([]domain.TaxPayer, error) {
	query := `
		SELECT id, identification, identification_type, name, email, COALESCE(address, ''), COALESCE(phone, ''), created_at, updated_at
		FROM tax_payers
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list taxpayers: %w", err)
	}
	defer rows.Close()

	var results []domain.TaxPayer
	for rows.Next() {
		var tp domain.TaxPayer
		if err := rows.Scan(&tp.ID, &tp.Identification, &tp.IdentificationType, &tp.Name, &tp.Email, &tp.Address, &tp.Phone, &tp.CreatedAt, &tp.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, tp)
	}
	return results, nil
}
