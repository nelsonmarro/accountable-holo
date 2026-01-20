package service

import (
	"context"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBudgetOverview(t *testing.T) {
	mockTxRepo := new(mocks.MockTransactionRepository)
	mockCatRepo := new(mocks.MockCategoryRepository)
	
	service := NewReportService(nil, mockTxRepo, mockCatRepo, nil, nil)

	ctx := context.Background()
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()
	
	t.Run("Returns correct budget status", func(t *testing.T) {
		// Arrange
		budgetVal := decimal.NewFromFloat(1000)
		categories := []domain.Category{
			{
				BaseEntity: domain.BaseEntity{ID: 1},
				Name:       "Suministros",
				Type:       domain.Outcome,
				MonthlyBudget: &budgetVal,
			},
			{
				BaseEntity: domain.BaseEntity{ID: 2},
				Name:       "Ventas",
				Type:       domain.Income,
				MonthlyBudget: &budgetVal, // Should be ignored as it is Income
			},
		}

		transactions := []domain.Transaction{
			{
				CategoryID: 1,
				Amount:     800.0,
				Category:   &categories[0],
			},
		}

		mockCatRepo.On("GetAllCategories", ctx).Return(categories, nil).Once()
		mockTxRepo.On("FindAllTransactions", ctx, mock.AnythingOfType("domain.TransactionFilters"), (*string)(nil)).
			Return(transactions, nil).Once()

		// Act
		statuses, err := service.GetBudgetOverview(ctx, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, statuses, 1)
		assert.Equal(t, "Suministros", statuses[0].CategoryName)
		assert.Equal(t, float64(80), statuses[0].PercentageUsed)
		assert.False(t, statuses[0].IsOverBudget)
		
		mockCatRepo.AssertExpectations(t)
		mockTxRepo.AssertExpectations(t)
	})

	t.Run("Identifies over budget", func(t *testing.T) {
		// Arrange
		budgetVal := decimal.NewFromFloat(500)
		categories := []domain.Category{
			{
				BaseEntity: domain.BaseEntity{ID: 1},
				Name:       "Suministros",
				Type:       domain.Outcome,
				MonthlyBudget: &budgetVal,
			},
		}

		transactions := []domain.Transaction{
			{
				CategoryID: 1,
				Amount:     600.0,
				Category:   &categories[0],
			},
		}

		mockCatRepo.On("GetAllCategories", ctx).Return(categories, nil).Once()
		mockTxRepo.On("FindAllTransactions", ctx, mock.AnythingOfType("domain.TransactionFilters"), (*string)(nil)).
			Return(transactions, nil).Once()

		// Act
		statuses, err := service.GetBudgetOverview(ctx, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.True(t, statuses[0].IsOverBudget)
		assert.Equal(t, float64(120), statuses[0].PercentageUsed)
	})
}