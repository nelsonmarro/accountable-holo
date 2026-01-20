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

func TestCreateUser(t *testing.T) {
	// Arrange: Clean the DB before the test
	truncateTables(t)
	ctx := context.Background()
	user := &domain.User{
		Username:     fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		PasswordHash: "password_hash",
		FirstName:    "Test",
		LastName:     "User",
		Role:         domain.CustomerRole,
	}

	// Act
	err := testUserRepo.CreateUser(ctx, user)

	// Assert
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestGetUserByUsername(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create a user first so we can fetch it
	createdUser := createTestUser(t, testUserRepo, "testuser", domain.CustomerRole)

	t.Run("should get an existing user", func(t *testing.T) {
		// Act
		foundUser, err := testUserRepo.GetUserByUsername(ctx, createdUser.Username)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, createdUser.Username, foundUser.Username)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		// Act
		foundUser, err := testUserRepo.GetUserByUsername(ctx, "nonexistent")

		// Assert
		require.Error(t, err)
		assert.Nil(t, foundUser)
	})
}

func TestGetUserByID(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create a user first so we can fetch it
	createdUser := createTestUser(t, testUserRepo, "testuser", domain.CustomerRole)

	t.Run("should get an existing user", func(t *testing.T) {
		// Act
		foundUser, err := testUserRepo.GetUserByID(ctx, createdUser.ID)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, createdUser.ID, foundUser.ID)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		// Act
		foundUser, err := testUserRepo.GetUserByID(ctx, 99999) // An ID that doesn't exist

		// Assert
		require.Error(t, err)
		assert.Nil(t, foundUser)
	})
}

func TestUpdateUser(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdUser := createTestUser(t, testUserRepo, "testuser", domain.CustomerRole)

	// Arrange: Modify the user details
	createdUser.Username = "updateduser"
	createdUser.FirstName = "Updated"
	createdUser.LastName = "User"
	createdUser.Role = domain.AdminRole
	originalUpdateTS := createdUser.UpdatedAt

	// Act
	time.Sleep(1 * time.Millisecond)
	err := testUserRepo.UpdateUser(ctx, createdUser)
	require.NoError(t, err)

	// Assert: Fetch the user again and check the new values
	updatedUser, err := testUserRepo.GetUserByID(ctx, createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, "updateduser", updatedUser.Username)
	assert.Equal(t, "Updated", updatedUser.FirstName)
	assert.Equal(t, "User", updatedUser.LastName)
	assert.Equal(t, domain.AdminRole, updatedUser.Role)
	assert.True(t, updatedUser.UpdatedAt.After(originalUpdateTS))
}

func TestDeleteUser(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	createdUser := createTestUser(t, testUserRepo, "testuser", domain.CustomerRole)

	// Act: Delete the user
	err := testUserRepo.DeleteUser(ctx, createdUser.ID)
	require.NoError(t, err)

	// Assert: Verify it's actually gone
	_, err = testUserRepo.GetUserByID(ctx, createdUser.ID)
	assert.Error(t, err, "Expected an error when getting a deleted user, but got none")
}

func TestGetAllUsers(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// Arrange: Create a couple of users
	createTestUser(t, testUserRepo, "user1", domain.CustomerRole)
	createTestUser(t, testUserRepo, "user2", domain.AdminRole)

	t.Run("should get all users", func(t *testing.T) {
		// Act
		users, err := testUserRepo.GetAllUsers(ctx)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, users)
		require.Len(t, users, 2)
	})
}
