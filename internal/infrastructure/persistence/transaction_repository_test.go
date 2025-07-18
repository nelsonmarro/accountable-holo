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
