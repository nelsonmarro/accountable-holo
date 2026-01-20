package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/shopspring/decimal"
)

// TransactionReportGenerator defines an interface for generating reports from a list of transactions.
type TransactionReportGenerator interface {
	SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string, currentUser *domain.User) error
}

// ReconciliationReportGenerator defines an interface for generating a reconciliation statement report.
type ReconciliationReportGenerator interface {
	ReconciliationStatementReport(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string, currentUser *domain.User) error
}

// DailyReportGenerator defines an interface for generating a daily financial report.
type DailyReportGenerator interface {
	DailyReport(ctx context.Context, report *domain.DailyReport, outputPath string, currentUser *domain.User) error
}

// ReportServiceImpl provides methods to generate financial reports.
type ReportServiceImpl struct {
	repo            ReportRepository
	transactionRepo TransactionRepository
	categoryRepo    CategoryRepository
	csvGenerator    interface { // <-- This is the
		TransactionReportGenerator
		DailyReportGenerator
	}
	pdfGenerator interface { // This generator must be able to handle all report types
		TransactionReportGenerator
		ReconciliationReportGenerator
		DailyReportGenerator
	}
}

// NewReportService creates a new instance of ReportServiceImpl.
func NewReportService(
	repo ReportRepository,
	transactionRepo TransactionRepository,
	categoryRepo CategoryRepository,
	csvGenerator interface {
		TransactionReportGenerator
		DailyReportGenerator
	},
	pdfGenerator interface {
		TransactionReportGenerator
		ReconciliationReportGenerator
		DailyReportGenerator
	},
) *ReportServiceImpl {
	return &ReportServiceImpl{
		repo:            repo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
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
	incomeMap := make(map[string]decimal.Decimal)
	expenseMap := make(map[string]decimal.Decimal)

	for _, tx := range transactions {
		if tx.Category != nil {
			amount := decimal.NewFromFloat(tx.Amount)
			switch tx.Category.Type {
			case domain.Income:
				totalIncome = totalIncome.Add(amount)
				incomeMap[tx.Category.Name] = incomeMap[tx.Category.Name].Add(amount)
			case domain.Outcome:
				totalExpenses = totalExpenses.Add(amount)
				expenseMap[tx.Category.Name] = expenseMap[tx.Category.Name].Add(amount)
			}
		}
	}

	netProfitLoss := totalIncome.Sub(totalExpenses)

	return domain.FinancialSummary{
		TotalIncome:        totalIncome,
		TotalExpenses:      totalExpenses,
		NetProfitLoss:      netProfitLoss,
		IncomeByCategory:   mapToSortedSlice(incomeMap),
		ExpensesByCategory: mapToSortedSlice(expenseMap),
	}, nil
}

func mapToSortedSlice(data map[string]decimal.Decimal) []domain.CategoryAmount {
	var result []domain.CategoryAmount
	for name, amount := range data {
		result = append(result, domain.CategoryAmount{
			CategoryName: name,
			Amount:       amount,
		})
	}
	// Sort by Amount descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Amount.GreaterThan(result[j].Amount)
	})
	return result
}

func (s *ReportServiceImpl) GenerateReportFile(ctx context.Context, format string, transactions []domain.Transaction, outputPath string, currentUser *domain.User) error {
	switch format {
	case "CSV":
		return s.csvGenerator.SelectedTransactionsReport(ctx, transactions, outputPath, currentUser)
	case "PDF":
		return s.pdfGenerator.SelectedTransactionsReport(ctx, transactions, outputPath, currentUser)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}
}

func (s *ReportServiceImpl) GenerateReconciliationReportFile(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string, currentUser *domain.User) error {
	return s.pdfGenerator.ReconciliationStatementReport(ctx, reconciliation, outputPath, currentUser)
}

func (s *ReportServiceImpl) GenerateDailyReportFile(ctx context.Context, report *domain.DailyReport, outputPath string, format string, currentUser *domain.User) error {
	switch format {
	case "CSV":
		return s.csvGenerator.DailyReport(ctx, report, outputPath, currentUser)
	case "PDF":
		return s.pdfGenerator.DailyReport(ctx, report, outputPath, currentUser)
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

func (s *ReportServiceImpl) GetBudgetOverview(
	ctx context.Context,
	startDate, endDate time.Time,
) ([]domain.BudgetStatus, error) {
	// 1. Get all categories
	categories, err := s.categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories for budget: %w", err)
	}

	// 2. Filter categories with budget
	var budgetedCategories []domain.Category
	for _, cat := range categories {
		if cat.MonthlyBudget != nil && !cat.MonthlyBudget.IsZero() && cat.Type == domain.Outcome {
			budgetedCategories = append(budgetedCategories, cat)
		}
	}

	if len(budgetedCategories) == 0 {
		return []domain.BudgetStatus{}, nil
	}

	// 3. Calculate "Months" in the range to adjust the budget
	days := endDate.Sub(startDate).Hours() / 24
	monthsFactor := decimal.NewFromFloat(1.0)
	if days > 32 {
		months := days / 30.0
		monthsFactor = decimal.NewFromFloat(months)
	}

	// 4. Calculate status for each
	var statuses []domain.BudgetStatus
	filters := domain.TransactionFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	transactions, err := s.transactionRepo.FindAllTransactions(ctx, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for budget: %w", err)
	}

	// Aggregate expenses by Category ID
	expensesMap := make(map[int]decimal.Decimal)
	for _, tx := range transactions {
		if tx.CategoryID != 0 && tx.Category != nil && tx.Category.Type == domain.Outcome {
			amount := decimal.NewFromFloat(tx.Amount)
			expensesMap[tx.CategoryID] = expensesMap[tx.CategoryID].Add(amount)
		}
	}

	for _, cat := range budgetedCategories {
		spent := expensesMap[cat.ID]
		adjustedBudget := cat.MonthlyBudget.Mul(monthsFactor)

		percentage := 0.0
		if !adjustedBudget.IsZero() {
			percentage, _ = spent.Div(adjustedBudget).Float64()
			percentage = percentage * 100
		}

		remaining := adjustedBudget.Sub(spent)
		isOver := spent.GreaterThan(adjustedBudget)

		statuses = append(statuses, domain.BudgetStatus{
			CategoryName:   cat.Name,
			BudgetAmount:   adjustedBudget,
			SpentAmount:    spent,
			PercentageUsed: percentage,
			Remaining:      remaining,
			IsOverBudget:   isOver,
		})
	}

	// Sort by Percentage Used Descending
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].PercentageUsed > statuses[j].PercentageUsed
	})

	return statuses, nil
}
