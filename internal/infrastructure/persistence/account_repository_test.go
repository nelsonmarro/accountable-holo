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

// Helper function to create a test account to avoid repetition
func createTestAccount(t *testing.T, repo *AccountRepositoryImpl) *domain.Account {
	acc := &domain.Account{
		Name:           "Test Bank Account",
		Number:         "12345",
		Type:           domain.SavingAccount,
		InitialBalance: 1000.50,
	}
	err := repo.CreateAccount(context.Background(), acc)
	require.NoError(t, err, "Failed to create test account")
	return acc
}

func createTestAccounts(t *testing.T, repo *AccountRepositoryImpl) []*domain.Account {
	acc1 := &domain.Account{
		Name:           "Test Bank Account 1",
		Number:         "12345",
		Type:           domain.SavingAccount,
		InitialBalance: 1000.50,
	}

	acc2 := &domain.Account{
		Name:           "Test Bank Account 2",
		Number:         "67890",
		Type:           domain.OrdinaryAccount,
		InitialBalance: 1000.50,
	}

	err := repo.CreateAccount(context.Background(), acc1)
	require.NoError(t, err, "Failed to create test account")

	err = repo.CreateAccount(context.Background(), acc2)
	require.NoError(t, err, "Failed to create test account")
	return []*domain.Account{acc1, acc2}
}

func TestCreateAccount(t *testing.T) {
	// Arrange: Clean the DB before the test
	truncateTables(t)
	ctx := context.Background()
	acc := &domain.Account{
		Name:           "Savings Account",
		Number:         "54321",
		Type:           domain.AccountType("savings"),
		InitialBalance: 50.25,
	}

	// Act
	err := testRepo.CreateAccount(ctx, acc)

	// Assert
	require.NoError(t, err)
	assert.NotZero(t, acc.ID) // Check that the ID was set by the DB
	assert.NotZero(t, acc.CreatedAt)
	assert.NotZero(t, acc.UpdatedAt)
}

func TestGetAllAccounts(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create an account first so we can fetch it
	createdAcc := createTestAccount(t, testRepo)

	t.Run("should get all accounts", func(t *testing.T) {
		// Act
		accounts, err := testRepo.GetAllAccounts(ctx)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, accounts)
		require.Len(t, accounts, 1) // We only created one account
		require.Equal(t, createdAcc.ID, accounts[0].ID)
	})
}

func TestGetAccountByID(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create an account first so we can fetch it
	createdAcc := createTestAccount(t, testRepo)

	t.Run("should get an existing account", func(t *testing.T) {
		// Act
		foundAcc, err := testRepo.GetAccountByID(ctx, createdAcc.ID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundAcc)
		assert.Equal(t, createdAcc.ID, foundAcc.ID)
		assert.Equal(t, createdAcc.Name, foundAcc.Name)
	})

	t.Run("should return error for non-existent account", func(t *testing.T) {
		// Act
		foundAcc, err := testRepo.GetAccountByID(ctx, 99999) // An ID that doesn't exist

		// Assert
		require.Error(t, err)
		assert.Nil(t, foundAcc)
	})
}

func TestDeleteAccount(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdAcc := createTestAccount(t, testRepo)

	// Act: Delete the account
	err := testRepo.DeleteAccount(ctx, createdAcc.ID)
	require.NoError(t, err)

	// Assert: Verify it's actually gone
	_, err = testRepo.GetAccountByID(ctx, createdAcc.ID)
	assert.Error(t, err, "Expected an error when getting a deleted account, but got none")
}

func TestUpdateAccount(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdAcc := createTestAccount(t, testRepo)

	// Arrange: Modify the account details
	createdAcc.Name = "Updated Account Name"
	createdAcc.Number = "98765"
	originalUpdateTS := createdAcc.UpdatedAt

	// Act
	// We need a small delay to ensure the updated_at timestamp changes
	time.Sleep(1 * time.Millisecond)
	err := testRepo.UpdateAccount(ctx, createdAcc)
	require.NoError(t, err)

	// Assert: Fetch the account again and check the new values
	updatedAcc, err := testRepo.GetAccountByID(ctx, createdAcc.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Account Name", updatedAcc.Name)
	assert.Equal(t, "98765", updatedAcc.Number)
	assert.True(t, updatedAcc.UpdatedAt.After(originalUpdateTS), "UpdatedAt timestamp should have been updated")
}

func TestAccountExistsForCreate(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	_ = createTestAccount(t, testRepo)

	t.Run("should return true for when creating and name exists on other account", func(t *testing.T) {
		exists, err := testRepo.AccountExists(ctx, "Test Bank Account", "943345", 0)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return true for when updating and number exists on other account", func(t *testing.T) {
		exists, err := testRepo.AccountExists(ctx, "Other Account", "12345", 0)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false when updating and no other accounts has the same name or number", func(t *testing.T) {
		exists, err := testRepo.AccountExists(ctx, "Non-Existent Account", "00000", 0)
		// Since we expect a false, an error here would be a problem
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestAccountExistsForUpdate(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	accounts := createTestAccounts(t, testRepo)

	t.Run("should return true for when updating and name exists on other account", func(t *testing.T) {
		exists, err := testRepo.AccountExists(ctx, "Test Bank Account 2", accounts[0].Number, accounts[0].ID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return true for when updating and number exists on other account", func(t *testing.T) {
		exists, err := testRepo.AccountExists(ctx, "Test Banck Account 3", "67890", accounts[0].ID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false when updating and no other accounts has the same name or number", func(t *testing.T) {
		exists, err := testRepo.AccountExists(ctx, "Non-Existent Account", "00000", accounts[0].ID)
		// Since we expect a false, an error here would be a problem
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
