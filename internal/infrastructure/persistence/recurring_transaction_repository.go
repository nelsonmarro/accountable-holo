package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/verith/internal/domain"
)

type RecurringTransactionRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewRecurringTransactionRepository(db *pgxpool.Pool) *RecurringTransactionRepositoryImpl {
	return &RecurringTransactionRepositoryImpl{db: db}
}

func (r *RecurringTransactionRepositoryImpl) Create(ctx context.Context, rt *domain.RecurringTransaction) error {
	query := `
		INSERT INTO recurring_transactions 
		(description, amount, account_id, category_id, interval, start_date, next_run_date, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		rt.Description,
		rt.Amount,
		rt.AccountID,
		rt.CategoryID,
		rt.Interval,
		rt.StartDate,
		rt.NextRunDate,
		rt.IsActive,
		now,
		now,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create recurring transaction: %w", err)
	}
	return nil
}

func (r *RecurringTransactionRepositoryImpl) GetAll(ctx context.Context) ([]domain.RecurringTransaction, error) {
	query := `
		SELECT id, description, amount, account_id, category_id, interval, start_date, next_run_date, is_active
		FROM recurring_transactions
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active recurring transactions: %w", err)
	}
	defer rows.Close()

	var results []domain.RecurringTransaction
	for rows.Next() {
		var rt domain.RecurringTransaction
		err := rows.Scan(
			&rt.ID, &rt.Description, &rt.Amount, &rt.AccountID, &rt.CategoryID,
			&rt.Interval, &rt.StartDate, &rt.NextRunDate, &rt.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring transaction: %w", err)
		}
		results = append(results, rt)
	}
	return results, nil
}

func (r *RecurringTransactionRepositoryImpl) GetAllActive(ctx context.Context) ([]domain.RecurringTransaction, error) {
	query := `
		SELECT id, description, amount, account_id, category_id, interval, start_date, next_run_date, is_active
		FROM recurring_transactions
		WHERE is_active = TRUE
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active recurring transactions: %w", err)
	}
	defer rows.Close()

	var results []domain.RecurringTransaction
	for rows.Next() {
		var rt domain.RecurringTransaction
		err := rows.Scan(
			&rt.ID, &rt.Description, &rt.Amount, &rt.AccountID, &rt.CategoryID,
			&rt.Interval, &rt.StartDate, &rt.NextRunDate, &rt.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recurring transaction: %w", err)
		}
		results = append(results, rt)
	}
	return results, nil
}

func (r *RecurringTransactionRepositoryImpl) Update(ctx context.Context, rt *domain.RecurringTransaction) error {
	query := `
		UPDATE recurring_transactions
		SET description = $1, amount = $2, interval = $3, start_date = $4, next_run_date = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.db.Exec(ctx, query,
		rt.Description,
		rt.Amount,
		rt.Interval,
		rt.StartDate,
		rt.NextRunDate,
		rt.IsActive,
		time.Now(),
		rt.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update recurring transaction: %w", err)
	}
	return nil
}

func (r *RecurringTransactionRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM recurring_transactions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}