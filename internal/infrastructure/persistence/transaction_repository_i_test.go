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

// Helper function to create a test transaction
func createTestTransaction(t *testing.T, txRepo *TransactionRepositoryImpl, accID, catID int, amount float64, date time.Time, userID int) *domain.Transaction {
	tx := &domain.Transaction{
		Description:     "Test Transaction",
		Amount:          amount,
		TransactionDate: date,
		AccountID:       accID,
		CategoryID:      catID,
		CreatedByID:     userID,
		UpdatedByID:     userID,
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
		user := createTestUser(t, testUserRepo, "testuser_update1", domain.AdminRole)
		acc := createTestAccount(t, accountRepo)
		cat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.0, time.Now(), user.ID)
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
		user := createTestUser(t, testUserRepo, "testuser_update2", domain.AdminRole)
		acc := createTestAccount(t, accountRepo)
		cat := createTestCategory(t, categoryRepo, "Salary", domain.Income)

		// Create a transaction in a past month to avoid future date errors
		lastMonth := time.Now().AddDate(0, -1, 0)
		tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.0, lastMonth, user.ID)
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
		user := createTestUser(t, testUserRepo, "testuser_update3", domain.AdminRole)
		acc := createTestAccount(t, accountRepo)
		incomeCat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		outcomeCat := createTestCategory(t, categoryRepo, "Rent", domain.Outcome)
		tx := createTestTransaction(t, txRepo, acc.ID, incomeCat.ID, 100.0, time.Now(), user.ID)
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
		user := createTestUser(t, testUserRepo, "testuser_update4", domain.AdminRole)
		acc := createTestAccount(t, accountRepo)
		cat := createTestCategory(t, categoryRepo, "Salary", domain.Income)
		_ = createTestCategory(t, categoryRepo, "Anular Transacción Ingreso", domain.Outcome)
		tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.0, time.Now(), user.ID)

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
	user := createTestUser(t, testUserRepo, "testuser_find", domain.AdminRole)
	acc := createTestAccount(t, accountRepo)
	catIncome := createTestCategory(t, categoryRepo, "Salary", domain.Income)
	catOutcome := createTestCategory(t, categoryRepo, "Groceries", domain.Outcome)
	catFreelance := createTestCategory(t, categoryRepo, "Freelance Work", domain.Income)

	// Create a set of transactions to test against
	now := time.Now().Truncate(time.Second)
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 2000.00, now.AddDate(0, 0, -10), user.ID)              // 1
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 50.00, now.AddDate(0, 0, -8), user.ID)                // 2
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 75.00, now.AddDate(0, 0, -6), user.ID)                // 3 - "Special Groceries"
	txToFind := createTestTransaction(t, txRepo, acc.ID, catFreelance.ID, 500.00, now.AddDate(0, 0, -5), user.ID) // 4
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 25.00, now.AddDate(0, 0, -3), user.ID)                // 5
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 2000.00, now.AddDate(0, 0, -1), user.ID)               // 6

	// Manually update one description for search testing
	txToFind.Description = "Special Freelance Project"
	err := txRepo.UpdateTransaction(ctx, txToFind)
	require.NoError(t, err)

	// --- Test Scenarios ---

	t.Run("should fetch all transactions with no filters", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 2, 4, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, nil)

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

	t.Run("should filter by search string", func(t *testing.T) {
		// Arrange
		search := "Special"
		filters := domain.TransactionFilters{}

		// Act
		result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, filters, &search)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, txToFind.ID, result.Data[0].ID)
	})
}

func TestFindAllTransactionsByAccount(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)
	ctx := context.Background()

	// --- Test Data Setup ---
	truncateTables(t) // Clear database before test
	user := createTestUser(t, testUserRepo, "testuser_findall_acc", domain.AdminRole)
	acc := createTestAccount(t, accountRepo)
	catIncome := createTestCategory(t, categoryRepo, "Salary", domain.Income)
	catOutcome := createTestCategory(t, categoryRepo, "Groceries", domain.Outcome)
	catFreelance := createTestCategory(t, categoryRepo, "Freelance Work", domain.Income)

	// Create a set of transactions to test against
	now := time.Now().Truncate(time.Second)
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 2000.00, now.AddDate(0, 0, -10), user.ID)
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 50.00, now.AddDate(0, 0, -8), user.ID)
	txToFind := createTestTransaction(t, txRepo, acc.ID, catFreelance.ID, 500.00, now.AddDate(0, 0, -5), user.ID)
	createTestTransaction(t, txRepo, acc.ID, catOutcome.ID, 25.00, now.AddDate(0, 0, -3), user.ID)
	createTestTransaction(t, txRepo, acc.ID, catIncome.ID, 2000.00, now.AddDate(0, 0, -1), user.ID)

	// Manually update one description for search testing
	txToFind.Description = "Special Freelance Project"
	err := txRepo.UpdateTransaction(ctx, txToFind)
	require.NoError(t, err)

	// --- Test Scenarios ---

	t.Run("should fetch all transactions with no filters", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}

		// Act
		result, err := txRepo.FindAllTransactionsByAccount(ctx, acc.ID, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 5, "Should find all 5 transactions")
	})

	t.Run("should filter by description", func(t *testing.T) {
		// Arrange
		desc := "Special Freelance"
		filters := domain.TransactionFilters{Description: &desc}

		// Act
		result, err := txRepo.FindAllTransactionsByAccount(ctx, acc.ID, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, txToFind.ID, result[0].ID)
	})

	t.Run("should filter by date range", func(t *testing.T) {
		// Arrange
		startDate := now.AddDate(0, 0, -8)
		endDate := now.AddDate(0, 0, -3)
		filters := domain.TransactionFilters{StartDate: &startDate, EndDate: &endDate}

		// Act
		result, err := txRepo.FindAllTransactionsByAccount(ctx, acc.ID, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 3, "Should find 3 transactions in the date range")
	})

	t.Run("should filter by category", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{CategoryID: &catOutcome.ID}

		// Act
		result, err := txRepo.FindAllTransactionsByAccount(ctx, acc.ID, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 2, "Should find both 'Groceries' transactions")
	})

	t.Run("should filter by transaction type", func(t *testing.T) {
		// Arrange
		incomeType := domain.Income
		filters := domain.TransactionFilters{CategoryType: &incomeType}

		// Act
		result, err := txRepo.FindAllTransactionsByAccount(ctx, acc.ID, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 3, "Should find all 3 income transactions")
		for _, tx := range result {
			require.NotNil(t, tx.Category)
			assert.Equal(t, domain.Income, tx.Category.Type)
		}
	})

	t.Run("should return no results for filters that don't match", func(t *testing.T) {
		// Arrange
		desc := "NonExistent"
		filters := domain.TransactionFilters{Description: &desc}

		// Act
		result, err := txRepo.FindAllTransactionsByAccount(ctx, acc.ID, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestFindAllTransactions(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)
	ctx := context.Background()

	// --- Test Data Setup ---
	truncateTables(t)                         // Clear database before test
	user := createTestUser(t, testUserRepo, "testuser_findall", domain.AdminRole)
	acc1 := createTestAccount(t, accountRepo) // acc1
	acc2 := createTestAccount(t, accountRepo) // acc2
	catIncome := createTestCategory(t, categoryRepo, "Salary", domain.Income)
	catOutcome := createTestCategory(t, categoryRepo, "Groceries", domain.Outcome)

	// Create transactions for both accounts
	now := time.Now().Truncate(time.Second)
	createTestTransaction(t, txRepo, acc1.ID, catIncome.ID, 1000, now.AddDate(0, 0, -5), user.ID)
	createTestTransaction(t, txRepo, acc1.ID, catOutcome.ID, 100, now.AddDate(0, 0, -4), user.ID)
	createTestTransaction(t, txRepo, acc2.ID, catIncome.ID, 2000, now.AddDate(0, 0, -3), user.ID)
	txToFind := createTestTransaction(t, txRepo, acc2.ID, catOutcome.ID, 200, now.AddDate(0, 0, -2), user.ID)

	// Manually update one description for search testing
	txToFind.Description = "Unique Project Description"
	err := txRepo.UpdateTransaction(ctx, txToFind)
	require.NoError(t, err)

	// --- Test Scenarios ---

	t.Run("should fetch all transactions from all accounts with no filters", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}

		// Act
		result, err := txRepo.FindAllTransactions(ctx, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 4, "Should find all 4 transactions from both accounts")
	})

	t.Run("should filter by description across all accounts", func(t *testing.T) {
		// Arrange
		desc := "Unique Project"
		filters := domain.TransactionFilters{Description: &desc}

		// Act
		result, err := txRepo.FindAllTransactions(ctx, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, txToFind.ID, result[0].ID)
	})

	t.Run("should filter by date range across all accounts", func(t *testing.T) {
		// Arrange
		startDate := now.AddDate(0, 0, -4)
		endDate := now.AddDate(0, 0, -3)
		filters := domain.TransactionFilters{StartDate: &startDate, EndDate: &endDate}

		// Act
		result, err := txRepo.FindAllTransactions(ctx, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 2, "Should find 2 transactions in the date range from both accounts")
	})

	t.Run("should filter by category across all accounts", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{CategoryID: &catOutcome.ID}

		// Act
		result, err := txRepo.FindAllTransactions(ctx, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 2, "Should find both 'Groceries' transactions from both accounts")
	})

	t.Run("should filter by transaction type across all accounts", func(t *testing.T) {
		// Arrange
		incomeType := domain.Income
		filters := domain.TransactionFilters{CategoryType: &incomeType}

		// Act
		result, err := txRepo.FindAllTransactions(ctx, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 2, "Should find all 2 income transactions from both accounts")
	})

	t.Run("should return no results for filters that don't match", func(t *testing.T) {
		// Arrange
		desc := "NonExistent"
		filters := domain.TransactionFilters{Description: &desc}

		// Act
		result, err := txRepo.FindAllTransactions(ctx, filters, nil)

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestFindTransactionsWithMultipleReceipts(t *testing.T) {
	// Setup Repositories
	accountRepo := NewAccountRepository(dbPool)
	categoryRepo := NewCategoryRepository(dbPool)
	txRepo := NewTransactionRepository(dbPool)
	ctx := context.Background()

	// --- Test Data Setup ---
	truncateTables(t)
	user := createTestUser(t, testUserRepo, "testuser_receipts", domain.AdminRole)
	acc := createTestAccount(t, accountRepo)
	cat := createTestCategory(t, categoryRepo, "Sales", domain.Income)

	// Create 1 transaction
	tx := createTestTransaction(t, txRepo, acc.ID, cat.ID, 100.00, time.Now(), user.ID)

	// Create Issuer
	var issuerID int
	err := dbPool.QueryRow(ctx, `
		INSERT INTO issuers (ruc, business_name, trade_name, establishment_address, main_address, establishment_code, emission_point_code, environment, keep_accounting, signature_path, created_at, updated_at)
		VALUES ('1790000000001', 'Test Issuer', 'Test Trade', 'Addr', 'Main Addr', '001', '001', 1, TRUE, '/tmp/dummy.p12', NOW(), NOW())
		RETURNING id
	`).Scan(&issuerID)
	require.NoError(t, err)

	// Create TaxPayer
	var taxPayerID int
	err = dbPool.QueryRow(ctx, `
		INSERT INTO tax_payers (identification, name, email, identification_type, created_at, updated_at)
		VALUES ('9999999999999', 'Consumer', 'test@test.com', '07', NOW(), NOW())
		RETURNING id
	`).Scan(&taxPayerID)
	require.NoError(t, err)

	// Insert MULTIPLE receipts for this transaction (simulating retries)
	
	// Receipt 1: Oldest, Failed
	_, err = dbPool.Exec(ctx, `
		INSERT INTO electronic_receipts (transaction_id, issuer_id, tax_payer_id, access_key, receipt_type, xml_content, sri_status, environment, created_at, updated_at)
		VALUES ($1, $2, $3, '1111111111111111111111111111111111111111111111111', '01', '<xml>old</xml>', 'RECHAZADA', 1, $4, $4)
	`, tx.ID, issuerID, taxPayerID, time.Now().Add(-2*time.Hour))
	require.NoError(t, err)

	// Receipt 2: New, Authorized
	_, err = dbPool.Exec(ctx, `
		INSERT INTO electronic_receipts (transaction_id, issuer_id, tax_payer_id, access_key, receipt_type, xml_content, sri_status, environment, created_at, updated_at)
		VALUES ($1, $2, $3, '2222222222222222222222222222222222222222222222222', '01', '<xml>new</xml>', 'AUTORIZADO', 1, $4, $4)
	`, tx.ID, issuerID, taxPayerID, time.Now())
	require.NoError(t, err)

	// --- Act ---
	result, err := txRepo.FindTransactionsByAccount(ctx, acc.ID, 1, 10, domain.TransactionFilters{}, nil)

	// --- Assert ---
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.TotalCount, "Should return exactly 1 transaction, ignoring duplicate receipts")
	assert.Len(t, result.Data, 1)
	
	returnedTx := result.Data[0]
	assert.Equal(t, tx.ID, returnedTx.ID)
	require.NotNil(t, returnedTx.ElectronicReceipt, "Should have receipt info")
	assert.Equal(t, "AUTORIZADO", returnedTx.ElectronicReceipt.SRIStatus, "Should show the status of the LATEST receipt")
	assert.Equal(t, "2222222222222222222222222222222222222222222222222", returnedTx.ElectronicReceipt.AccessKey)
}