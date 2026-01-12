package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransaction(t *testing.T) {
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockStorage := new(mocks.MockStorageService)
	// AccountService is not used in CreateTransaction logic currently
	mockAccService := new(mocks.MockAccountService)

	svc := service.NewTransactionService(mockTxRepo, mockStorage, mockAccService)
	ctx := context.Background()
	user := domain.User{BaseEntity: domain.BaseEntity{ID: 1}}

	t.Run("Success", func(t *testing.T) {
		tx := &domain.Transaction{
			AccountID:       1,
			CategoryID:      2,
			Amount:          100.0,
			TransactionDate: time.Now(),
			Description:     "Test Transaction",
		}

		mockTxRepo.On("CreateTransaction", ctx, tx).Return(nil).Once()

		err := svc.CreateTransaction(ctx, tx, user)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, tx.CreatedByID)
		mockTxRepo.AssertExpectations(t)
	})

	t.Run("Fail - Repository Error", func(t *testing.T) {
		tx := &domain.Transaction{
			AccountID:       99,
			CategoryID:      2,
			Amount:          100.0,
			TransactionDate: time.Now(),
			Description:     "Test",
		}
		
		// Simulate DB error (e.g., FK violation)
		mockTxRepo.On("CreateTransaction", ctx, tx).Return(errors.New("fk violation")).Once()

		err := svc.CreateTransaction(ctx, tx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al crear la transacción")
		mockTxRepo.AssertExpectations(t)
	})

	t.Run("Success - With Attachment", func(t *testing.T) {
		path := "file.pdf"
		tx := &domain.Transaction{
			AccountID:       1,
			CategoryID:      2,
			Amount:          100.0,
			TransactionDate: time.Now(),
			Description:     "Test",
			AttachmentPath:  &path,
		}

		mockStorage.On("Save", ctx, path, mock.Anything).Return("stored/file.pdf", nil).Once()
		mockTxRepo.On("UpdateAttachmentPath", ctx, mock.Anything, "stored/file.pdf").Return(nil).Once()
		
		// The CreateTransaction call happens before attachment logic
		mockTxRepo.On("CreateTransaction", ctx, tx).Return(nil).Once()

		err := svc.CreateTransaction(ctx, tx, user)
		assert.NoError(t, err)
	})
}

func TestVoidTransaction(t *testing.T) {
	mockTxRepo := new(mocks.MockTransactionRepository)
	svc := service.NewTransactionService(mockTxRepo, nil, nil)
	ctx := context.Background()
	user := domain.User{BaseEntity: domain.BaseEntity{ID: 1}}

	t.Run("Success", func(t *testing.T) {
		tx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: 10},
			IsVoided:   false,
		}

		// Expect check first
		mockTxRepo.On("GetTransactionByID", ctx, 10).Return(tx, nil).Once()
		// Then void
		mockTxRepo.On("VoidTransaction", ctx, 10, user).Return(nil).Once()

		err := svc.VoidTransaction(ctx, 10, user)
		assert.NoError(t, err)
	})

	t.Run("Fail - Already Voided", func(t *testing.T) {
		tx := &domain.Transaction{IsVoided: true}
		mockTxRepo.On("GetTransactionByID", ctx, 10).Return(tx, nil).Once()

		err := svc.VoidTransaction(ctx, 10, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ya ha sido anulada")
		mockTxRepo.AssertNotCalled(t, "VoidTransaction")
	})
}

func TestReconcileAccount(t *testing.T) {
	mockTxRepo := new(mocks.MockTransactionRepository)
	svc := service.NewTransactionService(mockTxRepo, nil, nil)
	ctx := context.Background()

	t.Run("Match", func(t *testing.T) {
		// System balance matches user input
		systemBalance := decimal.NewFromFloat(100.00)
		userBalance := decimal.NewFromFloat(100.00)
		endDate := time.Now()
		startDate := endDate.Add(-24 * time.Hour)

		mockTxRepo.On("GetBalanceAsOf", ctx, 1, startDate).Return(systemBalance, nil).Once()
		mockTxRepo.On("FindAllTransactionsByAccount", ctx, 1, mock.Anything, mock.Anything).Return([]domain.Transaction{}, nil).Once()

		recon, err := svc.ReconcileAccount(ctx, 1, startDate, endDate, userBalance)
		
		assert.NoError(t, err)
		assert.True(t, recon.Difference.IsZero())
	})

	t.Run("Discrepancy", func(t *testing.T) {
		systemBalance := decimal.NewFromFloat(100.00)
		userBalance := decimal.NewFromFloat(80.00) // User has less money
		endDate := time.Now()
		startDate := endDate.Add(-24 * time.Hour)

		mockTxRepo.On("GetBalanceAsOf", ctx, 1, startDate).Return(systemBalance, nil).Once()
		mockTxRepo.On("FindAllTransactionsByAccount", ctx, 1, mock.Anything, mock.Anything).Return([]domain.Transaction{}, nil).Once()

		recon, err := svc.ReconcileAccount(ctx, 1, startDate, endDate, userBalance)

		assert.NoError(t, err)
		// Difference = Actual - Calculated = 80 - 100 = -20
		assert.Equal(t, "-20", recon.Difference.String())
	})
}
