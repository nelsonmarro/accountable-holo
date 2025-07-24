package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/report"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

// ReportServiceImpl provides methods to generate financial reports.
type ReportServiceImpl struct {
	repo            ReportRepository
	transactionRepo TransactionRepository
	csvGenerator    report.ReportGenerator
	pdfGenerator    report.ReportGenerator
}

// NewReportService creates a new instance of ReportServiceImpl with the given repository.
func NewReportService(repo ReportRepository, transactionRepo TransactionRepository, csvGenerator report.ReportGenerator, pdfGenerator report.ReportGenerator) *ReportServiceImpl {
	return &ReportServiceImpl{repo: repo, transactionRepo: transactionRepo, csvGenerator: csvGenerator, pdfGenerator: pdfGenerator}
}

// GenerateFinancialSummary generates a financial summary report for the given date range.
func (s *ReportServiceImpl) GenerateFinancialSummary(ctx context.Context, startDate, endDate time.Time, accountID *int) (domain.FinancialSummary, error) {
	filters := domain.TransactionFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	transactions, err := s.transactionRepo.FindAllTransactionsByAccount(ctx, *accountID, 1, 1000000, filters) // Assuming a large enough page size for all transactions
	if err != nil {
		return domain.FinancialSummary{}, err
	}

	var totalIncome decimal.Decimal
	var totalExpenses decimal.Decimal

	for _, tx := range transactions {
		if tx.Category != nil {
			if tx.Category.Type == domain.Income {
				totalIncome = totalIncome.Add(decimal.NewFromFloat(tx.Amount))
			} else if tx.Category.Type == domain.Outcome {
				totalExpenses = totalExpenses.Add(decimal.NewFromFloat(tx.Amount))
			}
		}
	}

	netProfitLoss := totalIncome.Sub(totalExpenses)

	return domain.FinancialSummary{
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
		NetProfitLoss: netProfitLoss,
	}, nil
}

func (s *ReportServiceImpl) GenerateReportFile(ctx context.Context, format string, transactions []domain.Transaction, outputPath string) error {
	switch format {
	case "CSV":
		return s.csvGenerator.Generate(ctx, transactions, outputPath)
	case "PDF":
		return s.pdfGenerator.Generate(ctx, transactions, outputPath)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}
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
