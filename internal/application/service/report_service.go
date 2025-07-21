package service

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
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

