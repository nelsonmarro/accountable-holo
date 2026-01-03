package service

import (
	"context"
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFinancialSummary(t *testing.T) {
	mockTxRepo := new(mocks.MockTransactionRepository)
	
	// We don't need the other dependencies for this test
	service := NewReportService(nil, mockTxRepo, nil, nil, nil)

	ctx := context.Background()
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()
	
	t.Run("Calculates summary correctly for all accounts", func(t *testing.T) {
		// Arrange
		expectedIncome := 1000.0
		expectedExpense := 400.0
		
		transactions := []domain.Transaction{
			{
				Amount: expectedIncome,
				Category: &domain.Category{
					Type: domain.Income,
				},
			},
			{
				Amount: expectedExpense,
				Category: &domain.Category{
					Type: domain.Outcome,
				},
			},
		}

		// Expect FindAllTransactions (accountID is nil)
		mockTxRepo.On("FindAllTransactions", ctx, mock.AnythingOfType("domain.TransactionFilters"), (*string)(nil)).
			Return(transactions, nil).Once()

		// Act
		summary, err := service.GetFinancialSummary(ctx, startDate, endDate, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "1000", summary.TotalIncome.String())
		assert.Equal(t, "400", summary.TotalExpenses.String())
		assert.Equal(t, "600", summary.NetProfitLoss.String()) // 1000 - 400
		
		mockTxRepo.AssertExpectations(t)
	})

	t.Run("Calculates summary correctly for specific account", func(t *testing.T) {
		// Arrange
		accountID := 1
		expectedIncome := 500.50
		expectedExpense := 100.25
		
		transactions := []domain.Transaction{
			{
				Amount: expectedIncome,
				Category: &domain.Category{
					Type: domain.Income,
				},
			},
			{
				Amount: expectedExpense,
				Category: &domain.Category{
					Type: domain.Outcome,
				},
			},
		}

		// Expect FindAllTransactionsByAccount
		mockTxRepo.On("FindAllTransactionsByAccount", ctx, accountID, mock.AnythingOfType("domain.TransactionFilters"), (*string)(nil)).
			Return(transactions, nil).Once()

		// Act
		summary, err := service.GetFinancialSummary(ctx, startDate, endDate, &accountID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "500.5", summary.TotalIncome.String())
		assert.Equal(t, "100.25", summary.TotalExpenses.String())
		assert.Equal(t, "400.25", summary.NetProfitLoss.String()) // 500.50 - 100.25
		
		mockTxRepo.AssertExpectations(t)
	})
}
