//go:build integration

package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test category
func createTestCategory(t *testing.T, repo *CategoryRepositoryImpl, name string, catType domain.CategoryType) *domain.Category {
	cat := &domain.Category{
		Name: name,
		Type: catType,
	}
	err := repo.CreateCategory(context.Background(), cat)
	require.NoError(t, err, "Failed to create test category")
	return cat
}

// Helper function to create a test transaction
func createTestTransaction(t *testing.T, txRepo *TransactionRepositoryImpl, accID, catID int, amount float64, date time.Time) *domain.Transaction {
	tx := &domain.Transaction{
		Description:     "Test Transaction",
		Amount:          amount,
		TransactionDate: date,
		AccountID:       accID,
		CategoryID:      catID,
	}
	err := txRepo.CreateTransaction(context.Background(), tx)
	require.NoError(t, err, "Failed to create test transaction")
	return tx
}

func TestUpdateTransaction(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)

	// --- Test Scenarios ---
	t.Run("should update description without changing transaction number", func(t *testing.T) {
		// Arrange
		truncateTables(t)
		acc := createTestAccount(t, accountRepo)
		cat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.0, time.Now())
		originalTxNumber := tx.TransactionNumber

		// Act
		tx.Description = "Updated Description"
		err := txRepo.UpdateTransaction(context.Background(), tx)
		require.NoError(t, err)

		// Assert
		updatedTx, err := txRepo.GetTransactionByID(context.Background(), tx.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Description", updatedTx.Description)
		assert.Equal(t, originalTxNumber, updatedTx.TransactionNumber, "Transaction number should not have changed")
	})

	t.Run("should regenerate transaction number when date changes to a different month", func(t *testing.T) {
		// Arrange
		truncateTables(t)
		acc := createTestAccount(t, accountRepo)
		cat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		
		// Create a transaction in a past month to avoid future date errors
		lastMonth := time.Now().AddDate(0, -1, 0)
		tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.0, lastMonth)
		originalTxNumber := tx.TransactionNumber

		// Act
		// Move transaction to two months ago
		tx.TransactionDate = time.Now().AddDate(0, -2, 0)
		err := txRepo.UpdateTransaction(context.Background(), tx)
		require.NoError(t, err)

		// Assert
		updatedTx, err := txRepo.GetTransactionByID(context.Background(), tx.ID)
		require.NoError(t, err)
		assert.NotEqual(t, originalTxNumber, updatedTx.TransactionNumber, "Transaction number should have been regenerated")
		assert.Contains(t, updatedTx.TransactionNumber, tx.TransactionDate.Format("200601"), "Transaction number should contain the new month/year")
	})

	t.Run("should regenerate transaction number when category type changes", func(t *testing.T) {
		// Arrange
		truncateTables(t)
		acc := createTestAccount(t, accountRepo)
		incomeCat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		outcomeCat := createTestCategory(t, categoryRepo, "Rent", domain.Outcome)
		tx := createTestTransaction(t, txRepo, acc.ID, incomeCat.ID, 100.0, time.Now())
		originalTxNumber := tx.TransactionNumber
		require.Contains(t, originalTxNumber, "ING", "Initial transaction should be of type Income")

		// Act
		tx.CategoryID = outcomeCat.ID
		err := txRepo.UpdateTransaction(context.Background(), tx)
		require.NoError(t, err)

		// Assert
		updatedTx, err := txRepo.GetTransactionByID(context.Background(), tx.ID)
		require.NoError(t, err)
		assert.NotEqual(t, originalTxNumber, updatedTx.TransactionNumber, "Transaction number should have been regenerated")
		assert.Contains(t, updatedTx.TransactionNumber, "EGR", "New transaction number should have Outcome prefix")
	})

	t.Run("should fail to update a voided transaction", func(t *testing.T) {
		// Arrange
		truncateTables(t)
		acc := createTestAccount(t, accountRepo)
		cat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		_ = createTestCategory(t, categoryRepo, "Anular Transacción Ingreso", domain.Outcome)
		tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.0, time.Now())
		
		// Manually void the transaction for the test
		_, err := dbPool.Exec(context.Background(), "UPDATE transactions SET is_voided = TRUE WHERE id = $1", tx.ID)
		require.NoError(t, err)

		// Act
		tx.Description = "Attempted Update"
		err = txRepo.UpdateTransaction(context.Background(), tx)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no se puede actualizar una transacción previamente anulada")
	})
}

func TestFindTransactionsByAccount(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)
	ctx := context.Background()

	// --- Test Data Setup ---
	truncateTables(t) // Clear database before test
	acc := createTestAccount(t, accountRepo)
	catIncome := createTestCategory(t, categoryRepo, "Salary", domain.Income)
	catOutcome := createTestCategory(t, categoryRepo, "Groceries", domain.Outcome)
	catFreelance := createTestCategory(t, categoryRepo, "Freelance Work", domain.Income)

	// Create a set of transactions to test against
	now := time.Now().Truncate(time.Second)
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 2000.00, now.AddDate(0, 0, -10))      // 1
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 50.00, now.AddDate(0, 0, -8))       // 2
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 75.00, now.AddDate(0, 0, -6))       // 3 - "Special Groceries"
	txToFind := createTestTransaction(t, txRepo, acc.ID, catFreelance.ID, 500.00, now.AddDate(0, 0, -5)) // 4
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 25.00, now.AddDate(0, 0, -3))       // 5
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 2000.00, now.AddDate(0, 0, -1))      // 6
	
	// Manually update one description for search testing
	txToFind.Description = "Special Freelance Project"
	err := txRepo.UpdateTransaction(ctx, txToFind)
	require.NoError(t, err)


	// --- Test Scenarios ---

	t.Run("should fetch all transactions with no filters", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(6), result.TotalCount, "Should find all 6 transactions")
		assert.Len(t, result.Data, 6, "Data slice should contain 6 transactions")
		// Check if the most recent transaction is first
		assert.Equal(t, float64(2000.00), result.Data[0].Amount)
				assert.Equal(t, now.AddDate(0, 0, -1).Year(), result.Data[0].TransactionDate.Year())
		assert.Equal(t, now.AddDate(0, 0, -1).Month(), result.Data[0].TransactionDate.Month())
		assert.Equal(t, now.AddDate(0, 0, -1).Day(), result.Data[0].TransactionDate.Day())
	})

	t.Run("should filter by description", func(t *testing.T) {
		// Arrange
		desc := "Special Freelance"
		filters := domain.TransactionFilters{Description: &desc}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, txToFind.ID, result.Data[0].ID)
		assert.Contains(t, result.Data[0].Description, "Special Freelance Project")
	})

	t.Run("should filter by date range", func(t *testing.T) {
		// Arrange
		startDate := now.AddDate(0, 0, -8)
		endDate := now.AddDate(0, 0, -6)
		filters := domain.TransactionFilters{StartDate: &startDate, EndDate: &endDate}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount, "Should find 2 transactions in the date range")
		assert.Len(t, result.Data, 2)
		assert.Equal(t, float64(75.00), result.Data[0].Amount) // -6 days
		assert.Equal(t, float64(50.00), result.Data[1].Amount) // -8 days
	})

	t.Run("should filter by category", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{CategoryID: &catOutcome.ID}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.TotalCount, "Should find all 3 'Groceries' transactions")
		assert.Len(t, result.Data, 3)
		for _, tx := range result.Data {
			assert.Equal(t, catOutcome.ID, tx.CategoryID)
		}
	})

	t.Run("should filter by transaction type", func(t *testing.T) {
		// Arrange
		incomeType := domain.Income
		filters := domain.TransactionFilters{CategoryType: &incomeType}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.TotalCount, "Should find all 3 income transactions")
		assert.Len(t, result.Data, 3)
		for _, tx := range result.Data {
			require.NotNil(t, tx.Category)
			assert.Equal(t, domain.Income, tx.Category.Type)
		}
	})

	t.Run("should handle pagination correctly", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}

		// Act: Get the second page, with 4 items per page
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 2, 4, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(6), result.TotalCount, "Total count should still be 6")
		assert.Len(t, result.Data, 2, "The second page should have the remaining 2 transactions")
		// The 5th transaction overall is the first on the second page
		assert.Equal(t, float64(50.00), result.Data[0].Amount) 
		// The 6th transaction overall is the second on the second page
		assert.Equal(t, float64(2000.00), result.Data[1].Amount)
	})

	t.Run("should return no results for filters that don't match", func(t *testing.T) {
		// Arrange
		desc := "NonExistent"
		filters := domain.TransactionFilters{Description: &desc}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.TotalCount)
		assert.Len(t, result.Data, 0)
	})
	
	t.Run("should calculate running balance correctly", func(t *testing.T) {
		// Arrange
		// Account initial balance is 1000.50
		filters := domain.TransactionFilters{}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters)

		// Assert
		require.NoError(t, err)
		require.Len(t, result.Data, 6)

		// Balances are calculated from oldest to newest, but results are ordered newest to oldest.
		// Let's check a few key points in the results.
		// Initial Balance: 1000.50
		// 1. -10 days: +2000.00 -> 3000.50
		// 2. -8 days:  -50.00   -> 2950.50
		// 3. -6 days:  -75.00   -> 2875.50
		// 4. -5 days:  +500.00  -> 3375.50
		// 5. -3 days:  -25.00   -> 3350.50
		// 6. -1 day:   +2000.00 -> 5350.50

		// The most recent transaction (-1 day) should have the final balance
		assert.InDelta(t, 5350.50, result.Data[0].RunningBalance, 0.001) 
		
		// The transaction from -5 days ago should have its corresponding running balance
		var checkedTx domain.Transaction
		for _, tx := range result.Data {
			if tx.ID == txToFind.ID {
				checkedTx = tx
				break
			}
		}
		require.NotNil(t, checkedTx.ID, "Could not find the specific transaction to check balance")
		assert.InDelta(t, 3375.50, checkedTx.RunningBalance, 0.001, "Running balance for the -5 day transaction is incorrect")
	})
}