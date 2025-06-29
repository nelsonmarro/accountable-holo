package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TransactionRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{db: db}
}

func (r *TransactionRepositoryImpl) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		insert into transactions (description, amount, transaction_date, account_id, category_id, created_at, updated_at) 
											values ($1, $2, $3, $4, $5, $6, $7)
		                  returning id, created_at, updated_at`

	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		tx.Description,
		tx.Amount,
		tx.TransactionDate,
		tx.AccountID,
		tx.CategoryID,
		now,
		now,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

