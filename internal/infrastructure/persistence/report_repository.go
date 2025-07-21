package persistence

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// ReportRepositoryImpl implements the ReportRepository interface for generating financial reports.
type ReportRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewReportRepository creates a new instance of ReportRepositoryImpl with the provided database connection pool.
func NewReportRepository(db *pgxpool.Pool) *ReportRepositoryImpl {
	return &ReportRepositoryImpl{db: db}
}

func (r *ReportRepositoryImpl) GetFinancialSummary(ctx context.Context, startDate, endDate time.Time) (domain.FinancialSummary, error) {
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
	t.transaction_date >= $1 AND t.transaction_date <= $2;
    `
}
