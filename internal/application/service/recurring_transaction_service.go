package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type RecurringTransactionService struct {
	repo   RecurringTransactionRepository
	txRepo TransactionRepository
	logger *log.Logger
}

func NewRecurringTransactionService(repo RecurringTransactionRepository, txRepo TransactionRepository, logger *log.Logger) *RecurringTransactionService {
	return &RecurringTransactionService{
		repo:   repo,
		txRepo: txRepo,
		logger: logger,
	}
}

func (s *RecurringTransactionService) Create(ctx context.Context, rt *domain.RecurringTransaction) error {
	// Set initial next run date to the start date if not provided, or ensure logic consistency
	if rt.NextRunDate.IsZero() {
		rt.NextRunDate = rt.StartDate
	}
	rt.IsActive = true
	return s.repo.Create(ctx, rt)
}

func (s *RecurringTransactionService) GetAll(ctx context.Context) ([]domain.RecurringTransaction, error) {
	return s.repo.GetAll(ctx)
}

func (s *RecurringTransactionService) Update(ctx context.Context, rt *domain.RecurringTransaction) error {
	if rt.NextRunDate.IsZero() {
		// Calculate next run date if missing, though typically this comes from UI
		rt.NextRunDate = rt.StartDate
	}
	return s.repo.Update(ctx, rt)
}

func (s *RecurringTransactionService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// ProcessPendingRecurrences checks for due transactions and creates them.
// This should be called on app startup.
func (s *RecurringTransactionService) ProcessPendingRecurrences(ctx context.Context, systemUser domain.User) error {
	activeRecurrences, err := s.repo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active recurrences: %w", err)
	}

	today := time.Now()
	// Normalize today to start of day or end of day? End of day usually to catch everything due today.
	// But NextRunDate is a Date (00:00). So if NextRunDate <= Today, it's due.
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	count := 0
	for _, rt := range activeRecurrences {
		// While the next run date is in the past or today...
		// We loop because if the app wasn't opened for 2 months, we might need to generate 2 transactions.
		for !rt.NextRunDate.After(today) {
			// 1. Create the actual transaction
			tx := &domain.Transaction{
				Description:     fmt.Sprintf("%s (Recurrente)", rt.Description),
				Amount:          rt.Amount,
				TransactionDate: rt.NextRunDate,
				AccountID:       rt.AccountID,
				CategoryID:      rt.CategoryID,
				CreatedByUser:   &systemUser,
				// Note: We might need CreatedByID in domain if repo uses it directly
				CreatedByID: systemUser.ID,
				UpdatedByID: systemUser.ID,
			}

			if err := s.txRepo.CreateTransaction(ctx, tx); err != nil {
				s.logger.Printf("Error generating recurring transaction for ID %d: %v", rt.ID, err)
				break // Stop processing this recurrence to avoid infinite loop or bad state
			}

			// 2. Advance the NextRunDate
			nextDate := calculateNextDate(rt.NextRunDate, rt.Interval)
			rt.NextRunDate = nextDate

			// 3. Update the recurring definition
			if err := s.repo.Update(ctx, &rt); err != nil {
				s.logger.Printf("Error updating next run date for ID %d: %v", rt.ID, err)
				break
			}
			count++
		}
	}

	if count > 0 {
		s.logger.Printf("Successfully processed %d recurring transactions.", count)
	}
	return nil
}

func calculateNextDate(current time.Time, interval domain.RecurrenceInterval) time.Time {
	switch interval {
	case domain.IntervalMonthly:
		return current.AddDate(0, 1, 0)
	case domain.IntervalWeekly:
		return current.AddDate(0, 0, 7)
	default:
		return current.AddDate(0, 1, 0) // Default to monthly
	}
}
