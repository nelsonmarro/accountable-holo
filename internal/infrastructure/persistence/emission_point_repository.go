package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/verith/internal/domain"
)

type EmissionPointRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewEmissionPointRepository(db *pgxpool.Pool) *EmissionPointRepositoryImpl {
	return &EmissionPointRepositoryImpl{db: db}
}

func (r *EmissionPointRepositoryImpl) GetByPoint(ctx context.Context, issuerID int, estCode, pointCode, receiptType string) (*domain.EmissionPoint, error) {
	query := `
		SELECT id, issuer_id, establishment_code, emission_point_code, receipt_type, current_sequence, initial_sequence, is_active, created_at, updated_at
		FROM emission_points
		WHERE issuer_id = $1 AND establishment_code = $2 AND emission_point_code = $3 AND receipt_type = $4
	`
	var ep domain.EmissionPoint
	err := r.db.QueryRow(ctx, query, issuerID, estCode, pointCode, receiptType).Scan(
		&ep.ID,
		&ep.IssuerID,
		&ep.EstablishmentCode,
		&ep.EmissionPointCode,
		&ep.ReceiptType,
		&ep.CurrentSequence,
		&ep.InitialSequence,
		&ep.IsActive,
		&ep.CreatedAt,
		&ep.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get emission point: %w", err)
	}
	return &ep, nil
}

func (r *EmissionPointRepositoryImpl) Create(ctx context.Context, ep *domain.EmissionPoint) error {
	// Si es la primera vez y hay un secuencial inicial, el CurrentSequence empieza ahi menos 1
	// para que el primer IncrementSequence lo deje en el valor exacto.
	if ep.CurrentSequence == 0 && ep.InitialSequence > 0 {
		ep.CurrentSequence = ep.InitialSequence - 1
	}

	query := `
		INSERT INTO emission_points (issuer_id, establishment_code, emission_point_code, receipt_type, current_sequence, initial_sequence, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		ep.IssuerID,
		ep.EstablishmentCode,
		ep.EmissionPointCode,
		ep.ReceiptType,
		ep.CurrentSequence,
		ep.InitialSequence,
		ep.IsActive,
		now,
		now,
	).Scan(&ep.ID, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create emission point: %w", err)
	}
	return nil
}

func (r *EmissionPointRepositoryImpl) GetAllByIssuer(ctx context.Context, issuerID int) ([]domain.EmissionPoint, error) {
	query := `
		SELECT id, issuer_id, establishment_code, emission_point_code, receipt_type, current_sequence, initial_sequence, is_active, created_at, updated_at
		FROM emission_points
		WHERE issuer_id = $1
		ORDER BY receipt_type ASC, emission_point_code ASC
	`
	rows, err := r.db.Query(ctx, query, issuerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query emission points: %w", err)
	}
	defer rows.Close()

	var results []domain.EmissionPoint
	for rows.Next() {
		var ep domain.EmissionPoint
		err := rows.Scan(
			&ep.ID, &ep.IssuerID, &ep.EstablishmentCode, &ep.EmissionPointCode, &ep.ReceiptType,
			&ep.CurrentSequence, &ep.InitialSequence, &ep.IsActive, &ep.CreatedAt, &ep.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, ep)
	}
	return results, nil
}

func (r *EmissionPointRepositoryImpl) Update(ctx context.Context, ep *domain.EmissionPoint) error {
	query := `
		UPDATE emission_points
		SET current_sequence = $1, initial_sequence = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, ep.CurrentSequence, ep.InitialSequence, ep.IsActive, time.Now(), ep.ID)
	if err != nil {
		return fmt.Errorf("failed to update emission point: %w", err)
	}
	return nil
}

func (r *EmissionPointRepositoryImpl) IncrementSequence(ctx context.Context, id int) error {
	// Incrementamos de forma atómica en la base de datos
	query := `UPDATE emission_points SET current_sequence = current_sequence + 1, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment sequence: %w", err)
	}
	return nil
}
