package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Test Cases ----
func TestCreateNewAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("should create account succesfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockAccountRepository)
		accService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}

		// Tell the mock what to expect.
		mockRepo.On("AccountExists", ctx, acc.Number, acc.Number, 0).Return(false, nil)
		mockRepo.On("CreateAccount", ctx, acc).Return(nil)

		// Act
		err := accService.CreateNewAccount(ctx, acc)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should return error when fields are empty", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockAccountRepository) // The repo is fresh for each sub-test
		accountService := NewAccountService(mockRepo)
		acc := &domain.Account{} // empty account to trigger validation error

		// We DO NOT set up an expectation with .On() because we expect
		// the repository method to never be called due to the validation failure.

		// Act
		err := accountService.CreateNewAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "el campo Name es requerido")
		assert.Contains(t, err.Error(), "el campo InitialBalance es requerido")
		assert.Contains(t, err.Error(), "el campo Type es requerido")
		assert.Contains(t, err.Error(), "el campo Number es requerido")
		// Verify that the CreateAccount method was never called.
		mockRepo.AssertNotCalled(t, "CreateAccount")
	})

	t.Run("should return error when create account method in repo fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}
		repoError := errors.New("database connection failed")

		mockRepo.On("AccountExists", ctx, acc.Number, acc.Number, 0).Return(false, nil)
		mockRepo.On("CreateAccount", ctx, acc).Return(repoError)

		// Act
		err := accountService.CreateNewAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		// Check that our service wrapped the original error.
		assert.Contains(t, err.Error(), "failed to create account")
		assert.ErrorIs(t, err, repoError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when account exists method in repo fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}
		expecteErrorStr := "error al verificar si la cuenta existe"

		mockRepo.On("AccountExists", ctx, acc.Number, acc.Number, 0).Return(false, nil)

		// Act
		err := accountService.CreateNewAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		// Check that our service wrapped the original error.
		assert.Contains(t, err.Error(), expecteErrorStr)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "CreateNewAccount")
	})
}

func TestGetAccountByID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockAccountRepository)
	accountService := NewAccountService(mockRepo)

	t.Run("should return error for invalid ID", func(t *testing.T) {
		// Act
		_, err := accountService.GetAccountByID(ctx, 0)

		// Assert
		require.Error(t, err)
		assert.Equal(t, "invalid account ID", err.Error())
	})

	t.Run("should call repository for valid ID", func(t *testing.T) {
		// Arrange
		expectedAccount := &domain.Account{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Found Account"}
		testID := 1

		// Setup the expectation
		mockRepo.On("GetAccountByID", ctx, testID).Return(expectedAccount, nil)

		// Act
		resultAcc, err := accountService.GetAccountByID(ctx, testID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedAccount, resultAcc)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllAccounts(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockAccountRepository)
	accountService := NewAccountService(mockRepo)

	t.Run("should return a list of accounts", func(t *testing.T) {
		// Arrange
		testAccList := []domain.Account{
			{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Account 1"},
			{BaseEntity: domain.BaseEntity{ID: 2}, Name: "Account 2"},
		}
		// Setup the expectation
		mockRepo.On("GetAllAccounts", ctx).Return(testAccList, nil)

		// Act
		resultAccList, err := accountService.GetAllAccounts(ctx)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, testAccList, resultAccList)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateAccount(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.MockAccountRepository)
	accountService := NewAccountService(mockRepo)

	t.Run("should return error for invalid ID", func(t *testing.T) {
		// Act
		err := accountService.UpdateAccount(ctx, &domain.Account{BaseEntity: domain.BaseEntity{ID: 0}})

		// Assert
		require.Error(t, err)
		assert.Equal(t, "invalid account ID", err.Error())
	})

	t.Run("should return nil error if accound has ID", func(t *testing.T) {
		// Arrange
		updateAccount := &domain.Account{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Found Account"}

		// Setup the expectation
		mockRepo.On("UpdateAccount", ctx, updateAccount).Return(nil)

		// Act
		err := accountService.UpdateAccount(ctx, updateAccount)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error for invalid ID", func(t *testing.T) {
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)

		// No expectations needed as we're testing the service's input validation
		err := accountService.DeleteAccount(ctx, 0)

		require.Error(t, err)
		assert.Equal(t, "invalid account ID", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if the repo fail to delete account", func(t *testing.T) {
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)

		// Setup the expectation
		mockRepo.On("DeleteAccount", ctx, 1).Return(errors.New("sql error"))

		// Act
		err := accountService.DeleteAccount(ctx, 1)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete account:")
		require.Contains(t, err.Error(), "sql error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return nil error if a valid ID is passed", func(t *testing.T) {
		mockRepo := new(mocks.MockAccountRepository) // Create a new mock for this subtest
		accountService := NewAccountService(mockRepo)

		// Setup the expectation
		mockRepo.On("DeleteAccount", ctx, 1).Return(nil)

		// Act
		err := accountService.DeleteAccount(ctx, 1)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// ---- End of Test Cases ----
