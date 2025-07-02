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
	t.Run("should create account succesfully", func(t *testing.T) {
		ctx := context.Background()
		// Arrange
		mockRepo := new(mocks.MockAccountRepository)
		accService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}

		// Tell the mock what to expect.
		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, 0).Return(false, nil)
		mockRepo.On("CreateAccount", ctx, acc).Return(nil)

		// Act
		err := accService.CreateNewAccount(ctx, acc)

		// Assert
		mockRepo.AssertExpectations(t)
		require.NoError(t, err)
	})

	t.Run("should return error when fields are empty", func(t *testing.T) {
		ctx := context.Background()
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
		mockRepo.AssertNotCalled(t, "AccountExists")
	})

	t.Run("should return error when create account method in repo fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}
		repoError := errors.New("database connection failed")

		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, 0).Return(false, nil)
		mockRepo.On("CreateAccount", ctx, acc).Return(repoError)

		// Act
		err := accountService.CreateNewAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		// Check that our service wrapped the original error.
		assert.Contains(t, err.Error(), "error al crear la cuenta:")
		assert.ErrorIs(t, err, repoError)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when account exists method in repo fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}
		expecteErrStr := "error al verificar si la cuenta existe"
		expecteErr := errors.New("sql error")

		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, 0).Return(false, expecteErr)

		// Act
		err := accountService.CreateNewAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		// Check that our service wrapped the original error.
		assert.Contains(t, err.Error(), expecteErrStr)
		assert.Contains(t, err.Error(), expecteErr.Error())
		assert.ErrorIs(t, err, expecteErr)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "CreateNewAccount")
	})
	t.Run("should return error when account exists method in repo return true", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockRepo := new(mocks.MockAccountRepository)
		accountService := NewAccountService(mockRepo)
		acc := &domain.Account{Name: "Test Account", InitialBalance: 100, Number: "2222", Type: domain.SavingAccount}
		expecteErrStr := "ya existe una cuenta con el mismo nombre o número ingresado"

		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, 0).Return(true, nil)

		// Act
		err := accountService.CreateNewAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		// Check that our service wrapped the original error.
		assert.Contains(t, err.Error(), expecteErrStr)
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
		assert.Equal(t, "ID de cuenta inválido ", err.Error())
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
		assert.Equal(t, "ID de cuenta inválido ", err.Error())
		mockRepo.AssertNotCalled(t, "UpdateAccount")
		mockRepo.AssertNotCalled(t, "AccountExists")
	})

	t.Run("should return nil error if accound has ID", func(t *testing.T) {
		// Arrange
		acc := &domain.Account{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Found Account"}

		// Setup the expectation
		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, acc.ID).Return(false, nil)
		mockRepo.On("UpdateAccount", ctx, acc).Return(nil)

		// Act
		err := accountService.UpdateAccount(ctx, acc)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
	t.Run("should return error if account exists method in repo returns error", func(t *testing.T) {
		// Arrange
		acc := &domain.Account{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Found Account"}
		expecteErrStr := "error al verificar si la cuenta existe"
		expecteErr := errors.New("sql error")

		// Setup the expectation
		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, acc.ID).Return(false, expecteErr)

		// Act
		err := accountService.UpdateAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, expecteErr)
		require.Contains(t, err.Error(), expecteErrStr)
		require.Contains(t, err.Error(), expecteErr.Error())
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "UpdateAccount")
	})
	t.Run("should return error if account exists method in repo returns true", func(t *testing.T) {
		// Arrange
		acc := &domain.Account{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Found Account"}
		expecteErrStr := "ya existe otra cuenta con el mismo nombre o número ingresado "

		// Setup the expectation
		mockRepo.On("AccountExists", ctx, acc.Name, acc.Number, acc.ID).Return(true, nil)

		// Act
		err := accountService.UpdateAccount(ctx, acc)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), expecteErrStr)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "UpdateAccount")
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
		assert.Equal(t, "ID de cuenta inválido ", err.Error())
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "DeleteAccount")
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
		require.Contains(t, err.Error(), "error al eliminar la cuenta:")
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
