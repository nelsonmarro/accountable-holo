package service

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessPendingRecurrences(t *testing.T) {
	mockRecurRepo := new(mocks.MockRecurringTransactionRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)
	
	// Create a dummy logger that writes to nowhere (io.Discard) or just nil if handled
	logger := log.New(log.Writer(), "", 0) 

	service := NewRecurringTransactionService(mockRecurRepo, mockTxRepo, logger)

	ctx := context.Background()
	systemUser := domain.User{BaseEntity: domain.BaseEntity{ID: 1}}

	t.Run("Processes due monthly transaction", func(t *testing.T) {
		// Arrange
		lastRun := time.Now().AddDate(0, -1, -5) // 1 month and 5 days ago
		nextRun := lastRun.AddDate(0, 1, 0)      // Should have run 5 days ago

		recurrence := domain.RecurringTransaction{
			BaseEntity:  domain.BaseEntity{ID: 1},
			Description: "Alquiler",
			Amount:      1000,
			AccountID:   1,
			CategoryID:  2,
			Interval:    domain.IntervalMonthly,
			NextRunDate: nextRun,
			IsActive:    true,
		}

		mockRecurRepo.On("GetAllActive", ctx).Return([]domain.RecurringTransaction{recurrence}, nil).Once()
		
		// Expect transaction creation
		mockTxRepo.On("CreateTransaction", ctx, mock.MatchedBy(func(tx *domain.Transaction) bool {
			return tx.Amount == 1000 && tx.Description == "Alquiler (Recurrente)"
		})).Return(nil).Once()

		// Expect recurrence update (NextRunDate advanced)
		mockRecurRepo.On("Update", ctx, mock.MatchedBy(func(rt *domain.RecurringTransaction) bool {
			expectedNext := nextRun.AddDate(0, 1, 0)
			return rt.ID == 1 && rt.NextRunDate.Equal(expectedNext)
		})).Return(nil).Once()

		// Act
		err := service.ProcessPendingRecurrences(ctx, systemUser)

		// Assert
		assert.NoError(t, err)
		mockRecurRepo.AssertExpectations(t)
		mockTxRepo.AssertExpectations(t)
	})

	t.Run("Skips future transactions", func(t *testing.T) {
		// Arrange
		futureDate := time.Now().AddDate(0, 0, 5) // 5 days in future

		recurrence := domain.RecurringTransaction{
			BaseEntity:  domain.BaseEntity{ID: 2},
			NextRunDate: futureDate,
			IsActive:    true,
		}

		mockRecurRepo.On("GetAllActive", ctx).Return([]domain.RecurringTransaction{recurrence}, nil).Once()
		
		// Act
		err := service.ProcessPendingRecurrences(ctx, systemUser)

		// Assert
		assert.NoError(t, err)
		mockTxRepo.AssertNotCalled(t, "CreateTransaction")
		mockRecurRepo.AssertNotCalled(t, "Update")
	})
}

func TestCreateRecurringTransactionFromEdit(t *testing.T) {
	mockRecurRepo := new(mocks.MockRecurringTransactionRepository)
	mockTxRepo := new(mocks.MockTransactionRepository)
	logger := log.New(log.Writer(), "", 0)

	service := NewRecurringTransactionService(mockRecurRepo, mockTxRepo, logger)
	ctx := context.Background()

	t.Run("Create recurrence sets active flag and saves", func(t *testing.T) {
		// Simulate data coming from the Edit Dialog
		rt := &domain.RecurringTransaction{
			Description: "Pago Editado",
			Amount:      500,
			AccountID:   1,
			CategoryID:  3,
			Interval:    domain.IntervalMonthly,
			StartDate:   time.Now(),
			NextRunDate: time.Now().AddDate(0, 1, 0), // Next run is next month
		}

		// Expect repository Create call
		mockRecurRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.RecurringTransaction) bool {
			return r.IsActive == true && r.Description == "Pago Editado" && r.NextRunDate.After(time.Now())
		})).Return(nil).Once()

		// Act
		err := service.Create(ctx, rt)

		// Assert
		assert.NoError(t, err)
		mockRecurRepo.AssertExpectations(t)
	})
}