package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

type TransactionRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{db: db}
}

func (r *TransactionRepositoryImpl) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var cat domain.Category
	err = tx.QueryRow(ctx, "SELECT name, type FROM categories WHERE id = $1", transaction.CategoryID).
		Scan(&cat.Name, &cat.Type)
	if err != nil {
		return err
	}

	newTxNumber, err := r.generateTransactionNumber(ctx, tx, cat.Type, cat.Name, transaction.TransactionDate)
	if err != nil {
		return fmt.Errorf("failed to generate transaction number: %w", err)
	}

	transaction.TransactionNumber = newTxNumber

	query := `
		insert into transactions (transaction_number, description, amount, transaction_date, account_id, category_id, attachment_path, created_by_id, updated_by_id, created_at, updated_at)
				 values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				  returning id, created_at, updated_at`

	now := time.Now()
	err = tx.QueryRow(ctx, query,
		transaction.TransactionNumber,
		transaction.Description,
		transaction.Amount,
		transaction.TransactionDate,
		transaction.AccountID,
		transaction.CategoryID,
		transaction.AttachmentPath,
		transaction.CreatedByID,
		transaction.UpdatedByID,
		now,
		now,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *TransactionRepositoryImpl) FindTransactionsByAccount(
	ctx context.Context,
	accountID int,
	page int,
	pageSize int,
	filters domain.TransactionFilters,
	searchString *string,
) (*domain.PaginatedResult[domain.Transaction], error) {
	// --- Build the base query and arguments ---
	whereCondition, args := r.buildQueryConditions(filters, searchString, &accountID)

	// --- Get the total count for pagination ---
	countQuery := `
              SELECT COUNT(t.id)
              FROM transactions t
              LEFT JOIN categories c ON t.category_id = c.id
              WHERE ` + whereCondition

	var totalCount int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w",
			err)
	}

	if totalCount == 0 {
		return &domain.PaginatedResult[domain.Transaction]{
				Data:       []domain.Transaction{},
				TotalCount: 0,
				Page:       page,
				PageSize:   pageSize,
			},
			nil
	}

	// --- Build the main query for fetching the paginated data ---
	limit := pageSize
	offset := (page - 1) * pageSize
	paginationArgs := append(args, limit, offset)
	limitOffsetPlaceholders := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	finalQuery := fmt.Sprintf(`
    SELECT
        t.id,
        t.transaction_number,
        t.description,
        t.amount,
        t.transaction_date,
        t.account_id,
        t.category_id,
        t.attachment_path,
        t.is_voided,
        t.voided_by_transaction_id,
        t.voids_transaction_id,
        c.name AS category_name,
        c.type AS category_type,
		uc.username AS created_by,
		uu.username AS updated_by,
        (
            SELECT initial_balance FROM accounts WHERE id = t.account_id
        ) + (
            SELECT
                COALESCE(SUM(CASE WHEN c_inner.type = 'Ingreso' THEN t_inner.amount ELSE -t_inner.amount END), 0)
            FROM
                transactions AS t_inner
            JOIN
                categories AS c_inner ON t_inner.category_id = c_inner.id
            WHERE
                t_inner.account_id = t.account_id AND
                (t_inner.transaction_date < t.transaction_date OR (t_inner.transaction_date = t.transaction_date AND t_inner.id <= t.id))
        ) AS running_balance
    FROM
        transactions AS t
    LEFT JOIN
        categories AS c ON t.category_id = c.id
	LEFT JOIN
		users AS uc ON t.created_by_id = uc.id
	LEFT JOIN
		users AS uu ON t.updated_by_id = uu.id
    WHERE %s
    ORDER BY
        t.transaction_date DESC, t.id DESC
    %s;`, whereCondition, limitOffsetPlaceholders)

	rows, err := r.db.Query(ctx, finalQuery, paginationArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	transactions, err := r.scanTransactions(rows)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedResult[domain.Transaction]{
			Data:       transactions,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
		},
		nil
}

func (r *TransactionRepositoryImpl) FindAllTransactionsByAccount(
	ctx context.Context,
	accountID int,
	filters domain.TransactionFilters,
	searchString *string,
) ([]domain.Transaction, error) {
	// --- Build the base query and arguments ---
	whereCondition, args := r.buildQueryConditions(filters, searchString, &accountID)
	// --- Build the main query for fetching the paginated data ---
	finalQuery := fmt.Sprintf(`
    SELECT
        t.id,
        t.transaction_number,
        t.description,
        t.amount,
        t.transaction_date,
        t.account_id,
        t.category_id,
        t.attachment_path,
        t.is_voided,
        t.voided_by_transaction_id,
        t.voids_transaction_id,
        c.name AS category_name,
        c.type AS category_type,
		uc.username AS created_by,
		uu.username AS updated_by,
        (
            SELECT initial_balance FROM accounts WHERE id = t.account_id
        ) + (
            SELECT
                COALESCE(SUM(CASE WHEN c_inner.type = 'Ingreso' THEN t_inner.amount ELSE -t_inner.amount END), 0)
            FROM
                transactions AS t_inner
            JOIN
                categories AS c_inner ON t_inner.category_id = c_inner.id
            WHERE
                t_inner.account_id = t.account_id AND
                (t_inner.transaction_date < t.transaction_date OR (t_inner.transaction_date = t.transaction_date AND t_inner.id <= t.id))
        ) AS running_balance
    FROM
        transactions AS t
    LEFT JOIN
        categories AS c ON t.category_id = c.id
	LEFT JOIN
		users AS uc ON t.created_by_id = uc.id
	LEFT JOIN
		users AS uu ON t.updated_by_id = uu.id
    WHERE %s
    ORDER BY
        t.transaction_date DESC, t.id DESC;`, whereCondition)

	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *TransactionRepositoryImpl) FindAllTransactions(
	ctx context.Context,
	filters domain.TransactionFilters,
	searchString *string,
) ([]domain.Transaction, error) {
	// --- Build the base query and arguments ---
	whereCondition, args := r.buildQueryConditions(filters, searchString, nil)

	// --- Build the main query for fetching the paginated data ---
	finalQuery := fmt.Sprintf(`
    SELECT
        t.id,
        t.transaction_number,
        t.description,
        t.amount,
        t.transaction_date,
        t.account_id,
        t.category_id,
        t.attachment_path,
        t.is_voided,
        t.voided_by_transaction_id,
        t.voids_transaction_id,
        c.name AS category_name,
        c.type AS category_type,
			  uc.username AS created_by,
		    uu.username AS updated_by,
        (
            SELECT initial_balance FROM accounts WHERE id = t.account_id
        ) + (
            SELECT
                COALESCE(SUM(CASE WHEN c_inner.type = 'Ingreso' THEN t_inner.amount ELSE -t_inner.amount END), 0)
            FROM
                transactions AS t_inner
            JOIN
                categories AS c_inner ON t_inner.category_id = c_inner.id
            WHERE
                t_inner.account_id = t.account_id AND
                (t_inner.transaction_date < t.transaction_date OR (t_inner.transaction_date = t.transaction_date AND t_inner.id <= t.id))
        ) AS running_balance
    FROM
        transactions AS t
    LEFT JOIN
        categories AS c ON t.category_id = c.id
		LEFT JOIN
				users AS uc ON t.created_by_id = uc.id
		LEFT JOIN
				users AS uu ON t.updated_by_id = uu.id
    WHERE %s
    ORDER BY
        t.transaction_date DESC, t.id DESC;`, whereCondition)

	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *TransactionRepositoryImpl) GetBalanceAsOf(ctx context.Context, accountID int, date time.Time) (decimal.Decimal, error) {
	query := `
		       SELECT a.initial_balance + COALESCE(SUM(CASE WHEN type = 'Ingreso' THEN t.amount ELSE -t.amount END), 0)
           FROM accounts a
           LEFT JOIN transactions t ON a.id = t.account_id AND t.transaction_date < $2
           LEFT JOIN categories c ON t.category_id = c.id
           WHERE a.id = 1$
           GROUP BY a.initial_balance;
	`

	var balance decimal.Decimal
	err := r.db.QueryRow(ctx, query, accountID, date).Scan(&balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			var initialBalance decimal.Decimal
			err = r.db.QueryRow(ctx, "SELECT initial_balance FROM accounts WHERE id = $1", accountID).Scan(&initialBalance)
			if err != nil {
				return decimal.Zero, fmt.Errorf("failed to get initial balance for account %d: %w", accountID, err)
			}
			return initialBalance, nil
		}
		return decimal.Zero, fmt.Errorf("failed to get balance as of date: %w", err)
	}

	return balance, nil
}

func (r *TransactionRepositoryImpl) VoidTransaction(
	ctx context.Context,
	transactionID int,
	currentUser domain.User,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	originalTransactionQuery := `
		 SELECT
			 t.id,
			 t.transaction_number,
			 transaction_date,
			 t.description,
			 t.amount,
			 t.account_id,
			 t.is_voided,
			 t.voided_by_transaction_id,
			 t.voids_transaction_id,
			 c.type
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
		&originalTransaction.TransactionNumber,
		&originalTransaction.TransactionDate,
		&originalTransaction.Description,
		&originalTransaction.Amount,
		&originalTransaction.AccountID,
		&originalTransaction.IsVoided,
		&originalTransaction.VoidedByTransactionID,
		&originalTransaction.VoidsTransactionID,
		&originalCatType,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("transaction with ID %d not found", transactionID)
		}
		return fmt.Errorf("failed to get original transaction: %w", err)
	}

	if originalTransaction.IsVoided || originalTransaction.VoidsTransactionID != nil {
		return fmt.Errorf("no se puede anular una transacción previamente anulada o una transacción que anule a otra")
	}

	var opposingCatType domain.CategoryType
	if originalCatType == domain.Income {
		opposingCatType = domain.Outcome
	} else {
		opposingCatType = domain.Income
	}

	adjustmentCatQuery := `
		 select id, name
		 from categories
		where name like '%Anular Transacción%' and type = $1
	`

	var opposingCatID int
	var opposingCatName string
	err = tx.QueryRow(ctx, adjustmentCatQuery, opposingCatType).
		Scan(&opposingCatID, &opposingCatName)
	if err != nil {
		return fmt.Errorf("failed to get the opposing category: %w", err)
	}

	newDescription := "Anulación de la transacción #" + originalTransaction.TransactionNumber + ":\n" + originalTransaction.Description
	newTransactionDate := time.Now()

	voidTransactionNumber, err := r.generateTransactionNumber(ctx,
		tx,
		opposingCatType,
		opposingCatName,
		newTransactionDate,
	)
	if err != nil {
		return fmt.Errorf("failed to generate void transaction number: %w", err)
	}

	voidTransactionQuery := `
         insert into transactions
           (description, amount, transaction_date, account_id,
            category_id, voids_transaction_id, created_at, updated_at, transaction_number,
            created_by_id, updated_by_id)
            values1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id
     `

	var voidTransactionID int
	err = tx.QueryRow(
		ctx,
		voidTransactionQuery,
		newDescription,
		originalTransaction.Amount,
		newTransactionDate,
		originalTransaction.AccountID,
		opposingCatID,
		originalTransaction.ID,
		time.Now(),
		time.Now(),
		voidTransactionNumber,
		currentUser.ID,
		currentUser.ID,
	).Scan(&voidTransactionID)
	if err != nil {
		return fmt.Errorf("failed to create void transaction: %w", err)
	}

	// Mark the original transaction as voided
	markAsVoidedQuery := `
			update transactions set is_voided = TRUE, voided_by_transaction_id = $1, updated_by_id = $2
			 where id = $3
	`
	_, err = tx.Exec(ctx, markAsVoidedQuery, voidTransactionID, currentUser.ID, originalTransaction.ID)
	if err != nil {
		return fmt.Errorf("error when voiding the transaction: %d\nerror: %w", originalTransaction.ID, err)
	}

	// Mark that original transaction voids the new transaction
	markVoidsQuery := `
			update transactions set voids_transaction_id = $1
			where id = $2
	`

	_, err = tx.Exec(ctx, markVoidsQuery, originalTransaction.ID, voidTransactionID)
	if err != nil {
		return fmt.Errorf("error when assigning the voids_transaction_id on the new void transaction: %d\nerror: %w", voidTransactionID, err)
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
			select
				id,
				transaction_number,
				description,
				amount,
				transaction_date,
				account_id,
				category_id,
				attachment_path,
				is_voided,
				voided_by_transaction_id,
				voids_transaction_id,
				created_at,
				updated_at
			from transactions t
			   where id = $1
	`
	var tx domain.Transaction
	var attachment sql.NullString
	var voidedBy sql.NullInt64
	var voids sql.NullInt64

	err := r.db.QueryRow(ctx, query, transactionID).Scan(
		&tx.ID,
		&tx.TransactionNumber,
		&tx.Description,
		&tx.Amount,
		&tx.TransactionDate,
		&tx.AccountID,
		&tx.CategoryID,
		&attachment,
		&tx.IsVoided,
		&voidedBy,
		&voids,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("transaction with ID %d not found", transactionID)
		}
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	if attachment.Valid {
		tx.AttachmentPath = &attachment.String
	}
	if voidedBy.Valid {
		val := int(voidedBy.Int64)
		tx.VoidedByTransactionID = &val
	}
	if voids.Valid {
		val := int(voids.Int64)
		tx.VoidsTransactionID = &val
	}

	return &tx, nil
}

func (r *TransactionRepositoryImpl) UpdateTransaction(ctx context.Context, tx *domain.Transaction) error {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// Get original transaction to compare
	var originalTx domain.Transaction
	err = dbTx.QueryRow(ctx, `SELECT
		category_id,
		transaction_date,
		is_voided,
		voids_transaction_id
		FROM transactions WHERE id = $1`,
		tx.ID).
		Scan(
			&originalTx.CategoryID,
			&originalTx.TransactionDate,
			&originalTx.IsVoided,
			&originalTx.VoidsTransactionID,
		)
	if err != nil {
		return fmt.Errorf("failed to get original transaction data: %w", err)
	}

	if originalTx.IsVoided || originalTx.VoidsTransactionID != nil {
		return fmt.Errorf("no se puede actualizar una transacción previamente anulada o una transacción que anule a otra")
	}

	// Get new category info
	var newCat domain.Category
	err = dbTx.QueryRow(ctx, "SELECT name, type FROM categories WHERE id = $1", tx.CategoryID).
		Scan(&newCat.Name, &newCat.Type)
	if err != nil {
		return fmt.Errorf("failed to get new category info: %w", err)
	}

	// Get original category type
	var originalCatType domain.CategoryType
	err = dbTx.QueryRow(ctx, "SELECT type FROM categories WHERE id = $1", originalTx.CategoryID).
		Scan(&originalCatType)
	if err != nil {
		return fmt.Errorf("failed to get original category type: %w", err)
	}

	// Check if we need to regenerate the transaction number
	regenerateNumber := newCat.Type != originalCatType

	if tx.TransactionDate.Month() != originalTx.TransactionDate.Month() || tx.TransactionDate.Year() != originalTx.TransactionDate.Year() {
		regenerateNumber = true
	}

	if regenerateNumber {
		newTxNumber, err := r.generateTransactionNumber(ctx, dbTx, newCat.Type, newCat.Name, tx.TransactionDate)
		if err != nil {
			return fmt.Errorf("failed to generate new transaction number during update: %w", err)
		}
		tx.TransactionNumber = newTxNumber

		query := `
			UPDATE transactions
			SET description = $1, transaction_date = $2, category_id = $3, transaction_number = $4, attachment_path = $5, updated_at = $6, updated_by_id = $7
			WHERE id = $8
		`
		_, err = dbTx.Exec(ctx, query, tx.Description, tx.TransactionDate, tx.CategoryID, tx.TransactionNumber, tx.AttachmentPath, time.Now(), tx.UpdatedByID, tx.ID)
		if err != nil {
			return fmt.Errorf("failed to update transaction with new number: %w", err)
		}
	} else {
		// If no regeneration is needed, just update the relevant fields
		query := `
			UPDATE transactions
			SET description = $1, transaction_date = $2, category_id = $3, attachment_path = $4, updated_at = $5, updated_by_id = $6
			WHERE id = $7
		`
		_, err := dbTx.Exec(ctx, query, tx.Description, tx.TransactionDate, tx.CategoryID, tx.AttachmentPath, time.Now(), tx.UpdatedByID, tx.ID)
		if err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	return dbTx.Commit(ctx)
}

func (r *TransactionRepositoryImpl) UpdateAttachmentPath(ctx context.Context, transactionID int, attachmentPath string) error {
	query := `UPDATE transactions SET attachment_path = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, attachmentPath, time.Now(), transactionID)
	if err != nil {
		return fmt.Errorf("failed to update attachment path: %w", err)
	}
	return nil
}

func (r *TransactionRepositoryImpl) generateTransactionNumber(ctx context.Context, tx pgx.Tx, catType domain.CategoryType, catName string, date time.Time) (string, error) {
	var prefix string
	if catType == domain.Income {
		prefix = "ING"
	} else {
		prefix = "EGR"
	}
	if strings.Contains(catName, "Anular") {
		prefix = "ANU"
	}

	dateComp := date.Format("200601")
	sequenceQuery := `
		 SELECT COUNT(*) + 1
		  FROM transactions
		WHERE to_char(transaction_date, 'YYYYMM') = $1
	`
	var sequence int
	err := tx.QueryRow(ctx, sequenceQuery, dateComp).Scan(&sequence)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction sequence number: %w", err)
	}

	return fmt.Sprintf("%s-%s-%04d", prefix, dateComp, sequence), nil
}

func (r *TransactionRepositoryImpl) buildQueryConditions(
	filters domain.TransactionFilters,
	searchString *string,
	accountID *int,
) (string, []any) {
	// --- Build the base query and arguments ---
	args := []any{}
	whereClauses := []string{}
	argsCount := 1

	// --- Dynamically add WHERE clauses based on filters ---
	if accountID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.account_id = $%d", argsCount))
		args = append(args, *accountID)
		argsCount++
	}

	if filters.Description != nil && *filters.Description != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.description ILIKE $%d", argsCount))
		args = append(args, "%"+*filters.Description+"%")
		argsCount++
	}

	if filters.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.transaction_date >= $%d", argsCount))
		args = append(args, *filters.StartDate)
		argsCount++
	}

	if filters.EndDate != nil {
		endDate := *filters.EndDate
		nextDay := endDate.Add(24 * time.Hour)
		whereClauses = append(whereClauses, fmt.Sprintf("t.transaction_date < $%d", argsCount))
		args = append(args, nextDay)
		argsCount++
	}

	if filters.CategoryID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.category_id = $%d", argsCount))
		args = append(args, *filters.CategoryID)
		argsCount++
	}

	if filters.CategoryType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.type = $%d", argsCount))
		args = append(args, *filters.CategoryType)
		argsCount++
	}

	if searchString != nil && *searchString != "" {
		searchPattern := "%" + *searchString + "%"
		searchClauses := []string{}

		searchClauses = append(searchClauses, fmt.Sprintf("t.description ILIKE $%d", argsCount))
		args = append(args, searchPattern)
		argsCount++

		searchClauses = append(searchClauses, fmt.Sprintf("t.transaction_number ILIKE $%d", argsCount))
		args = append(args, searchPattern)
		argsCount++

		searchClauses = append(searchClauses, fmt.Sprintf("CAST(t.amount AS TEXT) ILIKE $%d", argsCount))
		args = append(args, searchPattern)
		argsCount++

		searchClauses = append(searchClauses, fmt.Sprintf("c.type ILIKE $%d", argsCount))
		args = append(args, searchPattern)
		argsCount++

		searchClauses = append(searchClauses, fmt.Sprintf("c.name ILIKE $%d", argsCount))
		args = append(args, searchPattern)
		argsCount++

		whereClauses = append(whereClauses, "("+strings.Join(searchClauses, " OR ")+")")
	}

	if len(whereClauses) == 0 {
		return "1 = 1", []any{}
	}

	return strings.Join(whereClauses, " AND "), args
}

func (r *TransactionRepositoryImpl) scanTransactions(rows pgx.Rows) ([]domain.Transaction, error) {
	transactions := make([]domain.Transaction, 0)
	for rows.Next() {
		var tx domain.Transaction
		var categoryName, categoryType, createdBy, updatedBy sql.NullString
		var attachment sql.NullString
		var voidedBy sql.NullInt64
		var voids sql.NullInt64
		err := rows.Scan(
			&tx.ID,
			&tx.TransactionNumber,
			&tx.Description,
			&tx.Amount,
			&tx.TransactionDate,
			&tx.AccountID,
			&tx.CategoryID,
			&attachment,
			&tx.IsVoided,
			&voidedBy,
			&voids,
			&categoryName,
			&categoryType,
			&createdBy,
			&updatedBy,
			&tx.RunningBalance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		if attachment.Valid {
			tx.AttachmentPath = &attachment.String
		}
		if voidedBy.Valid {
			val := int(voidedBy.Int64)
			tx.VoidedByTransactionID = &val
		}
		if voids.Valid {
			val := int(voids.Int64)
			tx.VoidsTransactionID = &val
		}
		if categoryName.Valid {
			tx.Category = &domain.Category{
				Name: categoryName.String,
				Type: domain.CategoryType(categoryType.String),
			}
		}
		if createdBy.Valid {
			tx.CreatedByUser = &domain.User{Username: createdBy.String}
		}
		if updatedBy.Valid {
			tx.UpdatedByUser = &domain.User{Username: updatedBy.String}
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over transactions: %w", err)
	}

	return transactions, nil
}
