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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}
	if page > 100 {
		page = 100 // Limit to 100 pages
	}

	offset := (page - 1) * pageSize

	// Base query
	query := `
		from transactions t
		join categories c on t.category_id = c.id
		where t.account_id = $1
	`
	args := []any{accountID}

	// Get total count
	countQuery := "select count(*)" + query
	var totalCount int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total transaction count: %w", err)
	}

	// Get paginated transactions
	dataQuery := `
		select
			t.id,
			t.description,
			t.amount,
			t.transaction_date,
			t.account_id,
			t.category_id,
			c.name as category_name,
			c.type as category_type
		` + query + `
		order by t.transaction_date desc
		limit $2 offset $3
	`
	args = append(args, pageSize, offset)

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

	// Calculate running balance
	// Get the balance up to the current page
	var balanceUpToPage float64
	balanceQuery := `
		select coalesce(sum(case when c.type = 'Ingreso' then t.amount else -t.amount end), 0)
		from transactions t
		join categories c on t.category_id = c.id
		where t.account_id = $1 and t.id not in (
			select id from transactions where account_id = $1 order by transaction_date desc limit $2
		)
	`
	err = r.db.QueryRow(ctx, balanceQuery, accountID, offset).Scan(&balanceUpToPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance up to page: %w", err)
	}

	// Add initial balance
	var initialBalance float64
	err = r.db.QueryRow(ctx, "select initial_balance from accounts where id = $1", accountID).Scan(&initialBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial balance: %w", err)
	}
	balanceUpToPage += initialBalance

	// Calculate running balance for the current page
	runningBalance := balanceUpToPage
	for i := range transactions {
		tx := &transactions[i]
		if tx.Category.Type == domain.Income {
			runningBalance += tx.Amount
		} else {
			runningBalance -= tx.Amount
		}
		tx.RunningBalance = runningBalance
	}

	return &domain.PaginatedResult[domain.Transaction]{
		Data:       transactions,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
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
		                   values($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = tx.Exec(
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
	)
	if err != nil {
		return fmt.Errorf("error when inserting the void transaction: %w", err)
	}

	// Mark the original transaction as voided
	markAsVoidedQuery := `
	  update transactions set is_voided = TRUE
		where id = $1
	`
	_, err = tx.Exec(ctx, markAsVoidedQuery, originalTransaction.ID)
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
