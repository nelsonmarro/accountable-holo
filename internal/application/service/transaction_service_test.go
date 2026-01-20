package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/application/service"
	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
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
		mockTxRepo.On("VoidTransaction", ctx, 10, user).Return(11, nil).Once()

		voidID, err := svc.VoidTransaction(ctx, 10, user)
		assert.NoError(t, err)
		assert.Equal(t, 11, voidID)
	})

	t.Run("Fail - Already Voided", func(t *testing.T) {
		tx := &domain.Transaction{IsVoided: true}
		mockTxRepo.On("GetTransactionByID", ctx, 10).Return(tx, nil).Once()

		voidID, err := svc.VoidTransaction(ctx, 10, user)
		assert.Error(t, err)
		assert.Equal(t, 0, voidID)
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

	t.Run("Complex Calculation", func(t *testing.T) {
		// Start with 1000
		startBalance := decimal.NewFromFloat(1000.00)
		startDate := time.Now().Add(-48 * time.Hour)
		endDate := time.Now()

		// Transactions in period: +200, -50, +300 = +450 net
		txs := []domain.Transaction{
			{Amount: 200, Category: &domain.Category{Type: domain.Income}},
			{Amount: 50, Category: &domain.Category{Type: domain.Outcome}},
			{Amount: 300, Category: &domain.Category{Type: domain.Income}},
		}

		mockTxRepo.On("GetBalanceAsOf", ctx, 1, startDate).Return(startBalance, nil).Once()
		mockTxRepo.On("FindAllTransactionsByAccount", ctx, 1, mock.Anything, mock.Anything).Return(txs, nil).Once()

		// Calculated should be 1000 + 450 = 1450
		// If user says they have 1400, discrepancy is -50
		actualBalance := decimal.NewFromFloat(1400.00)
		recon, err := svc.ReconcileAccount(ctx, 1, startDate, endDate, actualBalance)

		assert.NoError(t, err)
		assert.Equal(t, "1450", recon.CalculatedEndingBalance.String())
		assert.Equal(t, "-50", recon.Difference.String())
	})
}
