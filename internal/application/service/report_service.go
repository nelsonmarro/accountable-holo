package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

// TransactionReportGenerator defines an interface for generating reports from a list of transactions.
type TransactionReportGenerator interface {
	SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string) error
}

// ReconciliationReportGenerator defines an interface for generating a reconciliation statement report.
type ReconciliationReportGenerator interface {
	ReconciliationStatementReport(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string) error
}

// ReportServiceImpl provides methods to generate financial reports.
type ReportServiceImpl struct {
	repo            ReportRepository
	transactionRepo TransactionRepository
	csvGenerator    TransactionReportGenerator
	pdfGenerator    interface { // This generator must be able to handle both report types
		TransactionReportGenerator
		ReconciliationReportGenerator
	}
}

// NewReportService creates a new instance of ReportServiceImpl.
func NewReportService(
	repo ReportRepository,
	transactionRepo TransactionRepository,
	csvGenerator TransactionReportGenerator,
	pdfGenerator interface {
		TransactionReportGenerator
		ReconciliationReportGenerator
	},
) *ReportServiceImpl {
	return &ReportServiceImpl{
		repo:            repo,
		transactionRepo: transactionRepo,
		csvGenerator:    csvGenerator,
		pdfGenerator:    pdfGenerator,
	}
}

// GetFinancialSummary retrieves the financial summary for a given account within a date range.
func (s *ReportServiceImpl) GetFinancialSummary(
	ctx context.Context,
	startDate, endDate time.Time,
	accountID *int,
) (domain.FinancialSummary, error) {
	filters := domain.TransactionFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	var transactions []domain.Transaction
	var err error

	if accountID != nil {
		transactions, err = s.transactionRepo.FindAllTransactionsByAccount(ctx, *accountID, filters, nil)
	} else {
		transactions, err = s.transactionRepo.FindAllTransactions(ctx, filters, nil)
	}

	if err != nil {
		return domain.FinancialSummary{}, err
	}

	var totalIncome decimal.Decimal
	var totalExpenses decimal.Decimal

	for _, tx := range transactions {
		if tx.Category != nil {
			switch tx.Category.Type {
			case domain.Income:
				totalIncome = totalIncome.Add(decimal.NewFromFloat(tx.Amount))
			case domain.Outcome:
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
		return s.csvGenerator.SelectedTransactionsReport(ctx, transactions, outputPath)
	case "PDF":
		return s.pdfGenerator.SelectedTransactionsReport(ctx, transactions, outputPath)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}
}

func (s *ReportServiceImpl) GenerateReconciliationReportFile(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string) error {
	return s.pdfGenerator.ReconciliationStatementReport(ctx, reconciliation, outputPath)
}

func (s *ReportServiceImpl) GetReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time, endingBalance decimal.Decimal) (*domain.Reconciliation, error) {
	reconciliation, err := s.repo.GetReconciliation(ctx, accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	reconciliation.EndingBalance = endingBalance
	reconciliation.Difference = reconciliation.CalculatedEndingBalance.Sub(endingBalance)

	return reconciliation, nil
}
