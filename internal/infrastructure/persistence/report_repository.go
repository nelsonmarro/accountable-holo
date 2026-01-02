package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

// ReportRepositoryImpl implements the ReportRepository interface for generating financial reports.
type ReportRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewReportRepository creates a new instance of ReportRepositoryImpl with the provided database connection pool.
func NewReportRepository(db *pgxpool.Pool) *ReportRepositoryImpl {
	return &ReportRepositoryImpl{db: db}
}

// GetFinancialSummary retrieves a financial summary report for the specified date range, optionally filtered by account ID.
func (r *ReportRepositoryImpl) GetFinancialSummary(ctx context.Context, startDate, endDate time.Time, accountID *int) (domain.FinancialSummary, error) {
	var summary domain.FinancialSummary

	query := `
	SELECT
		COALESCE(SUM(CASE WHEN c.type = 'Ingreso' THEN t.amount ELSE 0 END), 0) AS total_income,
		COALESCE(SUM(CASE WHEN c.type = 'Egreso' THEN t.amount ELSE 0 END), 0) AS total_expenses
	FROM
		transactions t
	JOIN
		categories c ON t.category_id = c.id
	WHERE
		t.transaction_date >= $1 AND t.transaction_date <= $2`

	args := []interface{}{startDate, endDate}

	if accountID != nil {
		query += fmt.Sprintf(" AND t.account_id = $%d", len(args)+1)
		args = append(args, *accountID)
	}

	err := r.db.QueryRow(ctx, query, args...).Scan(&summary.TotalIncome, &summary.TotalExpenses)
	if err != nil {
		return summary, fmt.Errorf("failed to get financial summary: %w", err)
	}

	summary.NetProfitLoss = summary.TotalIncome.Sub(summary.TotalExpenses)

	return summary, nil
}

func (r *ReportRepositoryImpl) GetReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time) (*domain.Reconciliation, error) {
	query := `
	WITH initial_balance AS (
		SELECT 
			(SELECT initial_balance FROM accounts WHERE id = $1) + 
			COALESCE(SUM(CASE WHEN c.type = 'Ingreso' THEN t.amount ELSE -t.amount END), 0) as balance
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.account_id = $1 AND t.transaction_date < $2
	),
	transactions_in_period AS (
		SELECT t.*, c.name as category_name, c.type as category_type
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		WHERE t.account_id = $1 AND t.transaction_date >= $2 AND t.transaction_date <= $3
	)
	SELECT 
		ib.balance as starting_balance,
		COALESCE(SUM(CASE WHEN tip.category_type = 'Ingreso' THEN tip.amount ELSE -tip.amount END), 0) as net_movement,
		array_to_json(array_agg(row_to_json(tip)))
	FROM initial_balance ib
	LEFT JOIN transactions_in_period tip ON true
	GROUP BY ib.balance
	`

	rows, err := r.db.Query(ctx, query, accountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query for reconciliation: %w", err)
	}
	defer rows.Close()

	reconciliation := &domain.Reconciliation{
		AccountID:    accountID,
		StartDate:    startDate,
		EndDate:      endDate,
		Transactions: []domain.Transaction{},
	}

	if rows.Next() {
		var startingBalance float64
		var netMovement float64
		var transactionsJSON []byte

		if err := rows.Scan(&startingBalance, &netMovement, &transactionsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan reconciliation data: %w", err)
		}

		reconciliation.StartingBalance = decimal.NewFromFloat(startingBalance)
		reconciliation.CalculatedEndingBalance = decimal.NewFromFloat(startingBalance).Add(decimal.NewFromFloat(netMovement))

		if transactionsJSON != nil {
			var transactions []domain.Transaction
			if err := json.Unmarshal(transactionsJSON, &transactions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
			}
			reconciliation.Transactions = transactions
		}
	}

	return reconciliation, nil
}
