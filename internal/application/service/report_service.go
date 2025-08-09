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

// DailyReportGenerator defines an interface for generating a daily financial report.
type DailyReportGenerator interface {
	DailyReport(ctx context.Context, report *domain.DailyReport, outputPath string) error
}

// ReportServiceImpl provides methods to generate financial reports.
type ReportServiceImpl struct {
	repo            ReportRepository
	transactionRepo TransactionRepository
	csvGenerator    TransactionReportGenerator
	pdfGenerator    interface { // This generator must be able to handle all report types
		TransactionReportGenerator
		ReconciliationReportGenerator
		DailyReportGenerator
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
		DailyReportGenerator
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

func (s *ReportServiceImpl) GenerateDailyReportFile(ctx context.Context, report *domain.DailyReport, outputPath string, format string) error {
	switch format {
	case "CSV":
		// TODO: Implement CSV generation for daily report
		return fmt.Errorf("CSV format for daily report is not yet implemented")
	case "PDF":
		return s.pdfGenerator.DailyReport(ctx, report, outputPath)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}
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

func (s *ReportServiceImpl) GenerateDailyReport(ctx context.Context, accountID int) (*domain.DailyReport, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	balance, err := s.transactionRepo.GetBalanceAsOf(ctx, accountID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	filters := domain.TransactionFilters{
		StartDate: &startOfDay,
		EndDate:   &now,
	}
	transactions, err := s.transactionRepo.FindAllTransactionsByAccount(ctx, accountID, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily transactions: %w", err)
	}

	var dailyIncome, dailyExpenses decimal.Decimal
	for _, tx := range transactions {
		amount := decimal.NewFromFloat(tx.Amount)
		if tx.Category.Type == domain.Income {
			dailyIncome = dailyIncome.Add(amount)
		} else {
			dailyExpenses = dailyExpenses.Add(amount)
		}
	}

	return &domain.DailyReport{
		AccountID:       accountID,
		ReportDate:      now,
		CurrentBalance:  balance,
		DailyIncome:     dailyIncome,
		DailyExpenses:   dailyExpenses,
		DailyProfitLoss: dailyIncome.Sub(dailyExpenses),
		Transactions:    transactions,
	}, nil
}
