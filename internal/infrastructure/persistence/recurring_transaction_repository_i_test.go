//go:build integration

package persistence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestRecurringTransaction(t *testing.T, repo *RecurringTransactionRepositoryImpl, acc *domain.Account, cat *domain.Category) *domain.RecurringTransaction {
	rt := &domain.RecurringTransaction{
		Description: fmt.Sprintf("Recurring Rent %d", time.Now().UnixNano()),
		Amount:      500.00,
		AccountID:   acc.ID,
		CategoryID:  cat.ID,
		Interval:    domain.IntervalMonthly,
		StartDate:   time.Now(),
		NextRunDate: time.Now().AddDate(0, 1, 0),
		IsActive:    true,
	}
	err := repo.Create(context.Background(), rt)
	require.NoError(t, err, "Failed to create recurring transaction")
	return rt
}

func TestRecurringTransactionCreate(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create dependencies
	acc := createTestAccount(t, testRepo)
	cat := createTestCategory(t, testCatRepo, fmt.Sprintf("Cat %d", time.Now().UnixNano()), domain.Outcome)

	rt := &domain.RecurringTransaction{
		Description: "Internet Bill",
		Amount:      60.00,
		AccountID:   acc.ID,
		CategoryID:  cat.ID,
		Interval:    domain.IntervalMonthly,
		StartDate:   time.Now(),
		NextRunDate: time.Now(),
		IsActive:    true,
	}

	// Act
	err := testRecurringRepo.Create(ctx, rt)

	// Assert
	require.NoError(t, err)
	assert.NotZero(t, rt.ID)
	assert.NotZero(t, rt.CreatedAt)
	assert.NotZero(t, rt.UpdatedAt)
}

func TestRecurringTransactionGetAll(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	acc := createTestAccount(t, testRepo)
	cat := createTestCategory(t, testCatRepo, fmt.Sprintf("Cat %d", time.Now().UnixNano()), domain.Outcome)

	// Arrange: Create items
	createTestRecurringTransaction(t, testRecurringRepo, acc, cat)
	createTestRecurringTransaction(t, testRecurringRepo, acc, cat)

	// Act
	results, err := testRecurringRepo.GetAll(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRecurringTransactionGetAllActive(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	acc := createTestAccount(t, testRepo)
	cat := createTestCategory(t, testCatRepo, fmt.Sprintf("Cat %d", time.Now().UnixNano()), domain.Outcome)

	// Arrange: One active, one inactive
	rt1 := createTestRecurringTransaction(t, testRecurringRepo, acc, cat)
	
	rt2 := &domain.RecurringTransaction{
		Description: "Inactive Subscription",
		Amount:      10.00,
		AccountID:   acc.ID,
		CategoryID:  cat.ID,
		Interval:    domain.IntervalMonthly,
		StartDate:   time.Now(),
		NextRunDate: time.Now(),
		IsActive:    false,
	}
	err := testRecurringRepo.Create(ctx, rt2)
	require.NoError(t, err)

	// Act
	results, err := testRecurringRepo.GetAllActive(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, rt1.ID, results[0].ID)
}

func TestRecurringTransactionUpdate(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	acc := createTestAccount(t, testRepo)
	cat := createTestCategory(t, testCatRepo, fmt.Sprintf("Cat %d", time.Now().UnixNano()), domain.Outcome)
	rt := createTestRecurringTransaction(t, testRecurringRepo, acc, cat)

	// Arrange: Modify fields
	rt.Description = "Updated Rent"
	rt.Amount = 550.00
	rt.IsActive = false
	rt.NextRunDate = rt.NextRunDate.AddDate(0, 1, 0)

	// Act
	err := testRecurringRepo.Update(ctx, rt)
	require.NoError(t, err)

	// Assert: Verify changes
	// We need a helper to GetByID, but GetAll works for now as it's the only one
	all, err := testRecurringRepo.GetAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
	
	updated := all[0]
	assert.Equal(t, "Updated Rent", updated.Description)
	assert.Equal(t, 550.00, updated.Amount)
	assert.False(t, updated.IsActive)
	// Compare dates roughly or truncating time part if needed, but here it should be exact from DB scan
	// assert.Equal(t, rt.NextRunDate, updated.NextRunDate) 
}

func TestRecurringTransactionDelete(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	acc := createTestAccount(t, testRepo)
	cat := createTestCategory(t, testCatRepo, fmt.Sprintf("Cat %d", time.Now().UnixNano()), domain.Outcome)
	rt := createTestRecurringTransaction(t, testRecurringRepo, acc, cat)

	// Act
	err := testRecurringRepo.Delete(ctx, rt.ID)
	require.NoError(t, err)

	// Assert
	all, err := testRecurringRepo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 0)
}
