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

// Helper function to create a test category to avoid repetition
func createTestCategory(t *testing.T, repo *CategoryRepositoryImpl, name string, catType domain.CategoryType) *domain.Category {
	cat := &domain.Category{
		Name: name,
		Type: catType,
	}
	err := repo.CreateCategory(context.Background(), cat)
	require.NoError(t, err, "Failed to create test category")
	return cat
}

func TestCreateCategory(t *testing.T) {
	// Arrange: Clean the DB before the test
	truncateTables(t)
	ctx := context.Background()
	cat := &domain.Category{
		Name: fmt.Sprintf("Test Category %d", time.Now().UnixNano()),
		Type: domain.Income,
	}

	// Act
	err := testCatRepo.CreateCategory(ctx, cat)

	// Assert
	require.NoError(t, err)
	assert.NotZero(t, cat.ID)
	assert.NotZero(t, cat.CreatedAt)
	assert.NotZero(t, cat.UpdatedAt)
}

func TestGetAllCategories(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create a category first so we can fetch it
	createdCat := createTestCategory(t, testCatRepo, "Test Category", domain.Income)

	t.Run("should get all categories", func(t *testing.T) {
		// Act
		categories, err := testCatRepo.GetAllCategories(ctx)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, categories)
		require.Len(t, categories, 1)
		require.Equal(t, createdCat.ID, categories[0].ID)
	})
}

func TestGetCategoryByID(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create a category first so we can fetch it
	createdCat := createTestCategory(t, testCatRepo, "Test Category", domain.Income)

	t.Run("should get an existing category", func(t *testing.T) {
		// Act
		foundCat, err := testCatRepo.GetCategoryByID(ctx, createdCat.ID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundCat)
		assert.Equal(t, createdCat.ID, foundCat.ID)
		assert.Equal(t, createdCat.Name, foundCat.Name)
	})

	t.Run("should return error for non-existent category", func(t *testing.T) {
		// Act
		foundCat, err := testCatRepo.GetCategoryByID(ctx, 99999) // An ID that doesn't exist

		// Assert
		require.Error(t, err)
		assert.Nil(t, foundCat)
	})
}

func TestDeleteCategory(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdCat := createTestCategory(t, testCatRepo, "Test Category", domain.Income)

	// Act: Delete the category
	err := testCatRepo.DeleteCategory(ctx, createdCat.ID)
	require.NoError(t, err)

	// Assert: Verify it's actually gone
	_, err = testCatRepo.GetCategoryByID(ctx, createdCat.ID)
	assert.Error(t, err, "Expected an error when getting a deleted category, but got none")
}

func TestUpdateCategory(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdCat := createTestCategory(t, testCatRepo, "Test Category", domain.Income)

	// Arrange: Modify the category details
	createdCat.Name = "Updated Category Name"
	createdCat.Type = domain.Outcome
	originalUpdateTS := createdCat.UpdatedAt

	// Act
	// We need a small delay to ensure the updated_at timestamp changes
	time.Sleep(1 * time.Millisecond)
	err := testCatRepo.UpdateCategory(ctx, createdCat)
	require.NoError(t, err)

	// Assert: Fetch the category again and check the new values
	updatedCat, err := testCatRepo.GetCategoryByID(ctx, createdCat.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Category Name", updatedCat.Name)
	assert.Equal(t, domain.Outcome, updatedCat.Type)
	assert.True(t, updatedCat.UpdatedAt.After(originalUpdateTS), "UpdatedAt timestamp should have been updated")
}

func TestCategoryExists(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdCat := createTestCategory(t, testCatRepo, "Unique Name", domain.Income)

	t.Run("should return true when a category with the same name exists", func(t *testing.T) {
		exists, err := testCatRepo.CategoryExists(ctx, createdCat.Name, 0)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false when no category with that name exists", func(t *testing.T) {
		exists, err := testCatRepo.CategoryExists(ctx, "Non-Existent Name", 0)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should return false when checking for a name that belongs to the same category ID", func(t *testing.T) {
		exists, err := testCatRepo.CategoryExists(ctx, createdCat.Name, createdCat.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestFindByNameAndType(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdCat := createTestCategory(t, testCatRepo, "Findable Category", domain.Outcome)

	t.Run("should find an existing category by name and type", func(t *testing.T) {
		foundCat, err := testCatRepo.FindByNameAndType(ctx, "Findable Category", domain.Outcome)
		require.NoError(t, err)
		require.NotNil(t, foundCat)
		assert.Equal(t, createdCat.ID, foundCat.ID)
	})

	t.Run("should return error when category with name exists but type is different", func(t *testing.T) {
		foundCat, err := testCatRepo.FindByNameAndType(ctx, "Findable Category", domain.Income)
		require.Error(t, err)
		assert.Nil(t, foundCat)
	})

	t.Run("should return error for non-existent category", func(t *testing.T) {
		foundCat, err := testCatRepo.FindByNameAndType(ctx, "Non-Existent", domain.Outcome)
		require.Error(t, err)
		assert.Nil(t, foundCat)
	})
}
