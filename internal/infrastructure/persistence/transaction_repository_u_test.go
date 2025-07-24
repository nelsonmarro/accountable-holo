package persistence

import (
	"testing"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBuildQueryConditions(t *testing.T) {
	repo := &TransactionRepositoryImpl{}

	t.Run("should return 1=1 when no filters are provided", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}

		// Act
		where, args := repo.buildQueryConditions(filters, nil)

		// Assert
		assert.Equal(t, "1 = 1", where)
		assert.Empty(t, args)
	})

	t.Run("should build query with only account ID", func(t *testing.T) {
		// Arrange
		filters := domain.TransactionFilters{}
		accountID := 123

		// Act
		where, args := repo.buildQueryConditions(filters, &accountID)

		// Assert
		assert.Equal(t, "t.account_id = $1", where)
		assert.Equal(t, []any{123}, args)
	})

	t.Run("should build query with date range", func(t *testing.T) {
		// Arrange
		startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		filters := domain.TransactionFilters{
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		expectedEndDate := endDate.Add(24 * time.Hour) // The logic adds a day

		// Act
		where, args := repo.buildQueryConditions(filters, nil)

		// Assert
		assert.Equal(t, "t.transaction_date >= $1 AND t.transaction_date < $2", where)
		assert.Equal(t, []any{startDate, expectedEndDate}, args)
	})

	t.Run("should build query with all filters combined", func(t *testing.T) {
		// Arrange
		accountID := 456
		desc := "Grocery"
		catID := 789
		catType := domain.Outcome
		startDate := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC)

		filters := domain.TransactionFilters{
			Description:  &desc,
			StartDate:    &startDate,
			EndDate:      &endDate,
			CategoryID:   &catID,
			CategoryType: &catType,
		}
		expectedEndDate := endDate.Add(24 * time.Hour)

		// Act
		where, args := repo.buildQueryConditions(filters, &accountID)

		// Assert
		expectedWhere := "t.account_id = $1 AND t.description ILIKE $2 AND t.transaction_date >= $3 AND t.transaction_date < $4 AND t.category_id = $5 AND c.type = $6"
		assert.Equal(t, expectedWhere, where)

		expectedArgs := []any{456, "%Grocery%", startDate, expectedEndDate, 789, domain.Outcome}
		assert.Equal(t, expectedArgs, args)
	})
}
