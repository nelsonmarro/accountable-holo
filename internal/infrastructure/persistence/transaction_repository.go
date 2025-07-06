package persistence

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
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

func (r *TransactionRepositoryImpl) GetTransactionsByAccountPaginated(
	ctx context.Context,
	accountID,
	page,
	pageSize int,
	filter ...string,
) (*domain.PaginatedResult[domain.Transaction], error) {
	offset := (page - 1) * pageSize

	// TODO: Add filtering to the count query
	countQuery := `SELECT count(*) FROM transactions WHERE account_id = $1`
	var totalCount int64
	err := r.db.QueryRow(ctx, countQuery, accountID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total transaction count: %w", err)
	}

	// This query uses a window function to calculate the running balance in the database.
	// This is more efficient and less error-prone than calculating it in Go.
	dataQuery := `
		SELECT
			t.id,
			t.description,
			t.amount,
			t.transaction_date,
			t.account_id,
			t.category_id,
			c.name as category_name,
			c.type as category_type,
			a.initial_balance + SUM(CASE WHEN c.type = 'Ingreso' THEN t.amount ELSE -t.amount END) OVER (ORDER BY t.transaction_date ASC, t.id ASC) as running_balance
		FROM
			transactions t
		JOIN
			categories c ON t.category_id = c.id
		JOIN
			accounts a ON t.account_id = a.id
		WHERE
			t.account_id = $1
		ORDER BY
			t.transaction_date DESC, t.id DESC
		LIMIT $2 OFFSET $3
	`
	args := []any{accountID, pageSize, offset}

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]domain.Transaction, 0, pageSize)
	for rows.Next() {
		var tx domain.Transaction
		var categoryName string
		var categoryType domain.CategoryType
		err := rows.Scan(
			&tx.ID,
			&tx.Description,
			&tx.Amount,
			&tx.TransactionDate,
			&tx.AccountID,
			&tx.CategoryID,
			&categoryName,
			&categoryType,
			&tx.RunningBalance, // Scan the calculated running balance directly
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		tx.Category = &domain.Category{
			Name: categoryName,
			Type: categoryType,
		}
		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over transactions: %w", err)
	}

	return &domain.PaginatedResult[domain.Transaction]{
			Data:       transactions,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
		},
		nil
}

func (r *TransactionRepositoryImpl) VoidTransaction(ctx context.Context, transactionID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	originalTransactionQuery := `
	  SELECT t.id, t.description, t.amount, t.account_id, t.is_voided, c.type
	  FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.id = $1
		FOR UPDATE;
	`

	var originalTransaction domain.Transaction
	var originalCatType domain.CategoryType

	row := tx.QueryRow(ctx, originalTransactionQuery, transactionID)
	err = row.Scan(
		&originalTransaction.ID,
		&originalTransaction.Description,
		&originalTransaction.Amount,
		&originalTransaction.AccountID,
		&originalTransaction.IsVoided,
		&originalCatType,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("transaction with ID %d not found", transactionID)
		}
		return fmt.Errorf("failed to get original transaction: %w", err)
	}

	if originalTransaction.IsVoided {
		return fmt.Errorf("transaction with ID %d is already voided", transactionID)
	}

	var opposingCatType domain.CategoryType
	if originalCatType == domain.Income {
		opposingCatType = domain.Outcome
	} else {
		opposingCatType = domain.Income
	}

	adjustmentCatQuery := `
	  select id
	  from categoies
		where name like '%Anular Transacción%' and type = $1
	`

	var opposingCatID int
	err = tx.QueryRow(ctx, adjustmentCatQuery, opposingCatType).
		Scan(opposingCatID)
	if err != nil {
		return fmt.Errorf("failed to get the opposing category: %w", err)
	}

	newDescription := "Anulada por transaccion #" + strconv.Itoa(originalTransaction.ID) + ": " + originalTransaction.Description
	newTransactionDate := time.Now()

	voidTransactionQuery := `
	  insert into transactions
			                 (description, amount, transaction_date, account_id, category_id, voids_transaction_id, created_at, updated_at)
										 values($1, $2, $3, $4, $5, $6, $7, $8) returning id
	`

	var voidTransactionID int
	err = tx.QueryRow(
		ctx,
		voidTransactionQuery,
		&newDescription,
		&originalTransaction.Account,
		&newTransactionDate,
		&originalTransaction.AccountID,
		&opposingCatID,
		&originalTransaction.ID,
		time.Now(),
		time.Now(),
	).Scan(&voidTransactionID)
	if err != nil {
		return fmt.Errorf("error when inserting the void transaction: %w", err)
	}

	// Mark the original transaction as voided
	markAsVoidedQuery := `
			update transactions set is_voided = TRUE, voided_by_transaction_id = $1
		  where id = $2
	`
	_, err = tx.Exec(ctx, markAsVoidedQuery, voidTransactionID, originalTransaction.ID)
	if err != nil {
		return fmt.Errorf("error when voiding the transaction: %d\nerror: %w", originalTransaction.ID, err)
	}

	// Commit transactions
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepositoryImpl) GetTransactionByID(ctx context.Context, transactionID int) (*domain.Transaction, error) {
	query := `
			select id, description, amount, transaction_date, account_id, category_id, created_at, updated_at,
			from transactions t
      where id = $1
	`
	var tx domain.Transaction
	err := r.db.QueryRow(ctx, query, transactionID).Scan(
		&tx.ID,
		&tx.Description,
		&tx.Amount,
		&tx.TransactionDate,
		&tx.AccountID,
		&tx.CategoryID,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("transaction with ID %d not found", transactionID)
		}
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	return &tx, nil
}

func (r *TransactionRepositoryImpl) UpdateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		update transactions
		set description = $1, transaction_date = $2, category_id = $3, updated_at = $4
		where id = $6
	`
	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		tx.Description,
		tx.TransactionDate,
		tx.CategoryID,
		now,
		tx.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	return nil
}
