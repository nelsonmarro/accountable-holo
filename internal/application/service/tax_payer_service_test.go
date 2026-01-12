package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/application/service/mocks"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCreateTaxPayer(t *testing.T) {
	mockRepo := new(mocks.MockTaxPayerRepository)
	svc := service.NewTaxPayerService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification: "1790012345001",
			Name:           "Empresa Test",
			Email:          "test@email.com",
		}

		mockRepo.On("Create", ctx, tp).Return(nil).Once()

		err := svc.Create(ctx, tp)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Fail - Short Identification", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification: "123", // Invalid
			Name:           "Test",
			Email:          "test@email.com",
		}

		err := svc.Create(ctx, tp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "al menos 10 dígitos")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Fail - Missing Name", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification: "1790012345001",
			Name:           "", // Invalid
			Email:          "test@email.com",
		}

		err := svc.Create(ctx, tp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nombre y email son obligatorios")
	})

	t.Run("Fail - Missing Email", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification: "1790012345001",
			Name:           "Valid Name",
			Email:          "", // Invalid
		}

		err := svc.Create(ctx, tp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nombre y email son obligatorios")
	})

	t.Run("Fail - Repository Error", func(t *testing.T) {
		tp := &domain.TaxPayer{
			Identification: "1790012345001",
			Name:           "Empresa Test",
			Email:          "test@email.com",
		}

		mockRepo.On("Create", ctx, tp).Return(errors.New("db error")).Once()

		err := svc.Create(ctx, tp)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
	})
}

func TestGetTaxPayerByIdentification(t *testing.T) {
	mockRepo := new(mocks.MockTaxPayerRepository)
	svc := service.NewTaxPayerService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expected := &domain.TaxPayer{
			BaseEntity: domain.BaseEntity{ID: 1}, 
			Identification: "1790012345001",
		}
		
		mockRepo.On("GetByIdentification", ctx, "1790012345001").Return(expected, nil).Once()

		result, err := svc.GetByIdentification(ctx, "1790012345001")
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.On("GetByIdentification", ctx, "999").Return(nil, errors.New("not found")).Once()

		result, err := svc.GetByIdentification(ctx, "999")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSearchTaxPayer(t *testing.T) {
	mockRepo := new(mocks.MockTaxPayerRepository)
	svc := service.NewTaxPayerService(mockRepo)
	ctx := context.Background()

	t.Run("Success - Returns All", func(t *testing.T) {
		// En esta implementación simple, Search llama a GetAll
		expectedList := []domain.TaxPayer{
			{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Cliente A"},
			{BaseEntity: domain.BaseEntity{ID: 2}, Name: "Cliente B"},
		}

		mockRepo.On("GetAll", ctx).Return(expectedList, nil).Once()

		result, err := svc.Search(ctx, "cualquier query")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedList, result)
	})

	t.Run("Fail - Repo Error", func(t *testing.T) {
		mockRepo.On("GetAll", ctx).Return(nil, errors.New("db fail")).Once()

		result, err := svc.Search(ctx, "query")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
