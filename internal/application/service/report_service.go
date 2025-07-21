package service

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

// ReportServiceImpl provides methods to generate financial reports.
type ReportServiceImpl struct {
	repo ReportRepository
}

// ReportRepository defines the methods required for report generation.
func NewReportService(repo ReportRepository) *ReportServiceImpl {
	return &ReportServiceImpl{repo: repo}
}

// GenerateFinancialSummary generates a financial summary report for the given date range.
func (s *ReportServiceImpl) GenerateFinancialSummary(ctx context.Context, startDate, endDate time.Time) (domain.FinancialSummary, error) {
	return s.repo.GetFinancialSummary(ctx, startDate, endDate)
}

func (s *ReportServiceImpl) GenerateReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time, endingBalance decimal.Decimal) (*domain.Reconciliation, error) {
	reconciliation, err := s.repo.GetReconciliation(ctx, accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	reconciliation.EndingBalance = endingBalance
	reconciliation.Difference = reconciliation.CalculatedEndingBalance.Sub(endingBalance)

	return reconciliation, nil
}
