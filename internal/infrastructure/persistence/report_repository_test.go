//go:build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFinancialSummary(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)
	reportRepo := NewReportRepository(dbPool)
	ctx := context.Background()

	// --- Test Data Setup ---
	truncateTables(t) // Clear database before test
	acc1 := createTestAccount(t, accountRepo)
	acc2 := createTestAccount(t, accountRepo)
	catIncome := createTestCategory(t, categoryRepo, "Salary", domain.Income)
	catOutcome := createTestCategory(t, categoryRepo, "Groceries", domain.Outcome)

	// Create transactions for both accounts
	now := time.Now().Truncate(time.Second)
	createTestTransaction(t, txRepo, acc1.ID, catIncome.ID, 1000, now.AddDate(0, 0, -5))
	createTestTransaction(t, txRepo, acc1.ID, catOutcome.ID, 100, now.AddDate(0, 0, -4))
	createTestTransaction(t, txRepo, acc2.ID, catIncome.ID, 2000, now.AddDate(0, 0, -3))
	createTestTransaction(t, txRepo, acc2.ID, catOutcome.ID, 200, now.AddDate(0, 0, -2))

	// --- Test Scenarios ---
	t.Run("should get financial summary for all accounts", func(t *testing.T) {
		// Arrange
		startDate := now.AddDate(0, 0, -10)
		endDate := now

		// Act
		summary, err := reportRepo.GetFinancialSummary(ctx, startDate, endDate, nil)

		// Assert
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(3000).Equal(summary.TotalIncome), "Total income should be 3000")
		assert.True(t, decimal.NewFromFloat(300).Equal(summary.TotalExpenses), "Total expenses should be 300")
		assert.True(t, decimal.NewFromFloat(2700).Equal(summary.NetProfitLoss), "Net profit/loss should be 2700")
	})

	t.Run("should get financial summary for a single account", func(t *testing.T) {
		// Arrange
		startDate := now.AddDate(0, 0, -10)
		endDate := now

		// Act
		summary, err := reportRepo.GetFinancialSummary(ctx, startDate, endDate, &acc1.ID)

		// Assert
		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(1000).Equal(summary.TotalIncome), "Total income for acc1 should be 1000")
		assert.True(t, decimal.NewFromFloat(100).Equal(summary.TotalExpenses), "Total expenses for acc1 should be 100")
		assert.True(t, decimal.NewFromFloat(900).Equal(summary.NetProfitLoss), "Net profit/loss for acc1 should be 900")
	})
}

func TestGetReconciliation(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)
	reportRepo := NewReportRepository(dbPool)
	ctx := context.Background()

	// --- Test Data Setup ---
	truncateTables(t) // Clear database before test
	acc := createTestAccount(t, accountRepo)
	catIncome := createTestCategory(t, categoryRepo, "Salary", domain.Income)
	catOutcome := createTestCategory(t, categoryRepo, "Groceries", domain.Outcome)

	// Create transactions for the account
	now := time.Now().Truncate(time.Second)
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 1000, now.AddDate(0, 0, -10)) // Before period
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 50, now.AddDate(0, 0, -9))   // Before period
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 500, now.AddDate(0, 0, -5))   // In period
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 100, now.AddDate(0, 0, -4))  // In period
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 200, now.AddDate(0, 0, -1))   // In period

	// --- Test Scenarios ---
	t.Run("should get reconciliation data correctly", func(t *testing.T) {
		// Arrange
		startDate := now.AddDate(0, 0, -8)
		endDate := now

		// Act
		reconciliation, err := reportRepo.GetReconciliation(ctx, acc.ID, startDate, endDate)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, reconciliation)

		// Initial balance is 1000.50 (from createTestAccount) + 1000 - 50 = 1950.50
		assert.True(t, decimal.NewFromFloat(1950.50).Equal(reconciliation.StartingBalance), "Starting balance should be 1950.50")

		// Net movement is 500 - 100 + 200 = 600
		// Ending balance is 1950.50 + 600 = 2550.50
		assert.True(t, decimal.NewFromFloat(2550.50).Equal(reconciliation.CalculatedEndingBalance), "Calculated ending balance should be 2550.50")

		assert.Len(t, reconciliation.Transactions, 3, "Should have 3 transactions in the period")
	})
}
