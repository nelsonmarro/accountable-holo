package service

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

func TestRecurringTransactionService_ProcessPendingRecurrences(t *testing.T) {
	ctx := context.Background()
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	systemUser := domain.User{BaseEntity: domain.BaseEntity{ID: 1}, Username: "system"}

	t.Run("No pending recurrences", func(t *testing.T) {
		repo := new(mocks.MockRecurringTransactionRepository)
		txRepo := new(mocks.MockTransactionRepository)
		svc := NewRecurringTransactionService(repo, txRepo, logger)

		// Setup: One future recurrence
		futureDate := time.Now().AddDate(0, 0, 1)
		recurrences := []domain.RecurringTransaction{
			{
				BaseEntity:  domain.BaseEntity{ID: 1},
				Description: "Future Internet",
				Amount:      50.0,
				NextRunDate: futureDate,
				IsActive:    true,
			},
		}

		repo.On("GetAllActive", ctx).Return(recurrences, nil)

		err := svc.ProcessPendingRecurrences(ctx, systemUser)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		txRepo.AssertNotCalled(t, "CreateTransaction", mock.Anything, mock.Anything)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("One pending monthly recurrence (due today)", func(t *testing.T) {
		repo := new(mocks.MockRecurringTransactionRepository)
		txRepo := new(mocks.MockTransactionRepository)
		svc := NewRecurringTransactionService(repo, txRepo, logger)

		today := time.Now()
		// Normalize to start of day
		today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

		recurrence := domain.RecurringTransaction{
			BaseEntity:  domain.BaseEntity{ID: 1},
			Description: "Rent",
			Amount:      1000.0,
			Interval:    domain.IntervalMonthly,
			NextRunDate: today,
			IsActive:    true,
		}

		repo.On("GetAllActive", ctx).Return([]domain.RecurringTransaction{recurrence}, nil)
		
		// Expect one transaction creation
		txRepo.On("CreateTransaction", ctx, mock.MatchedBy(func(tx *domain.Transaction) bool {
			return tx.Amount == 1000.0 && tx.TransactionDate.Equal(today)
		})).Return(nil)

		// Expect update to next month
		nextMonth := today.AddDate(0, 1, 0)
		repo.On("Update", ctx, mock.MatchedBy(func(rt *domain.RecurringTransaction) bool {
			return rt.NextRunDate.Equal(nextMonth)
		})).Return(nil)

		err := svc.ProcessPendingRecurrences(ctx, systemUser)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		txRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("Catch-up: Multiple pending weekly recurrences", func(t *testing.T) {
		repo := new(mocks.MockRecurringTransactionRepository)
		txRepo := new(mocks.MockTransactionRepository)
		svc := NewRecurringTransactionService(repo, txRepo, logger)

		// App wasn't opened for 3 weeks
		threeWeeksAgo := time.Now().AddDate(0, 0, -21)
		threeWeeksAgo = time.Date(threeWeeksAgo.Year(), threeWeeksAgo.Month(), threeWeeksAgo.Day(), 0, 0, 0, 0, threeWeeksAgo.Location())

		recurrence := domain.RecurringTransaction{
			BaseEntity:  domain.BaseEntity{ID: 1},
			Description: "Weekly Water",
			Amount:      10.0,
			Interval:    domain.IntervalWeekly,
			NextRunDate: threeWeeksAgo,
			IsActive:    true,
		}

		repo.On("GetAllActive", ctx).Return([]domain.RecurringTransaction{recurrence}, nil)
		
		// Expect 4 creations (3 weeks ago, 2 weeks ago, 1 week ago, today)
		// 21 / 7 = 3 intervals past + today = 4?
		// Wait: 21, 14, 7, 0 days ago. Yes, 4 executions.
		txRepo.On("CreateTransaction", ctx, mock.Anything).Return(nil).Times(4)
		repo.On("Update", ctx, mock.Anything).Return(nil).Times(4)

		err := svc.ProcessPendingRecurrences(ctx, systemUser)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		txRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}
