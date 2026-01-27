//go:build integration

package persistence

import (
	"context"
	"testing"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaxPayerRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Use global dbPool initialized in TestMain
	repo := NewTaxPayerRepository(dbPool)
	ctx := context.Background()
	
	// Clean slate
	truncateTables(t)

	t.Run("Create and GetByIdentification", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification:     "1790012345001",
			IdentificationType: "04",
			Name:               "Test Company S.A.",
			Email:              "test@company.com",
			Address:            "Av. Amazonas y Naciones Unidas",
			Phone:              "0991234567",
		}

		err := repo.Create(ctx, tp)
		require.NoError(t, err)
		assert.NotZero(t, tp.ID)
		assert.NotZero(t, tp.CreatedAt) // Sufficient for DB persistence check

		// Fetch back
		fetched, err := repo.GetByIdentification(ctx, "1790012345001")
		require.NoError(t, err)
		assert.NotNil(t, fetched)
		assert.Equal(t, tp.ID, fetched.ID)
		assert.Equal(t, "Test Company S.A.", fetched.Name)
	})

	t.Run("GetByID", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification:     "1104567890",
			IdentificationType: "05",
			Name:               "Juan Perez",
			Email:              "juan@perez.com",
		}
		err := repo.Create(ctx, tp)
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, tp.Identification, fetched.Identification)
	})

	t.Run("GetAll", func(t *testing.T) {
		// Clean table first if needed, or count existing
		initialList, err := repo.GetAll(ctx)
		require.NoError(t, err)
		initialCount := len(initialList)

		// Add 2 more
		repo.Create(ctx, &domain.TaxPayer{Identification: "9999999999999", Name: "Final", Email: "a@a.com"})
		repo.Create(ctx, &domain.TaxPayer{Identification: "1234567890", Name: "Otro", Email: "b@b.com"})

		finalList, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, initialCount+2, len(finalList))
	})

	t.Run("Update", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification: "5555555555",
			Name:           "Original Name",
			Email:          "original@mail.com",
		}
		repo.Create(ctx, tp)

		tp.Name = "Updated Name"
		tp.Address = "New Address"
		err := repo.Update(ctx, tp)
		require.NoError(t, err)

		fetched, _ := repo.GetByID(ctx, tp.ID)
		assert.Equal(t, "Updated Name", fetched.Name)
		assert.Equal(t, "New Address", fetched.Address)
	})

	t.Run("GetPaginated", func(t *testing.T) {
		truncateTables(t)

		// Seed 5 taxpayers
		repo.Create(ctx, &domain.TaxPayer{Identification: "1000000001", Name: "Alpha", Email: "a@test.com"})
		repo.Create(ctx, &domain.TaxPayer{Identification: "1000000002", Name: "Beta", Email: "b@test.com"})
		repo.Create(ctx, &domain.TaxPayer{Identification: "1000000003", Name: "Gamma", Email: "g@test.com"})
		repo.Create(ctx, &domain.TaxPayer{Identification: "1000000004", Name: "Delta", Email: "d@test.com"})
		repo.Create(ctx, &domain.TaxPayer{Identification: "1000000005", Name: "Epsilon", Email: "e@test.com"})

		// Test Page 1, Size 2
		result, err := repo.GetPaginated(ctx, 1, 2, "")
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.TotalCount)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, "Alpha", result.Data[0].Name)
		assert.Equal(t, "Beta", result.Data[1].Name)

		// Test Page 3, Size 2 (Last page)
		result, err = repo.GetPaginated(ctx, 3, 2, "")
		require.NoError(t, err)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, "Gamma", result.Data[0].Name)

		// Test Search
		result, err = repo.GetPaginated(ctx, 1, 10, "del")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "Delta", result.Data[0].Name)
	})
}
