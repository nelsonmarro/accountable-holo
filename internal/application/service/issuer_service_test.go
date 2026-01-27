package service

import (
	"context"
	"testing"

	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zalando/go-keyring"
)

func TestSaveIssuerConfig_Integration(t *testing.T) {
	// Usar el backend de memoria de keyring para testing
	keyring.MockInit()

	mockRepo := new(mocks.MockIssuerRepository)
	mockEpRepo := new(mocks.MockEmissionPointRepository)
	service := NewIssuerService(mockRepo, mockEpRepo)
	ctx := context.Background()

	t.Run("Create New Issuer Config", func(t *testing.T) {
		// Arrange
		newIssuer := &domain.Issuer{
			RUC:          "1790012345001",
			BusinessName: "Mi Empresa S.A.",
			IsActive:     true,
		}
		password := "SecretPass123!"

		// Expects
		mockRepo.On("GetActive", ctx).Return(nil, nil).Once() // No existe
		mockRepo.On("Create", ctx, newIssuer).Return(nil).Once()
		
		// New logic calls GetByPoint and Create for default points (01, 04)
		mockEpRepo.On("GetByPoint", ctx, mock.Anything, mock.Anything, mock.Anything, "01").Return(nil, nil).Once()
		mockEpRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
		mockEpRepo.On("GetByPoint", ctx, mock.Anything, mock.Anything, mock.Anything, "04").Return(nil, nil).Once()
		mockEpRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		// Act
		err := service.SaveIssuerConfig(ctx, newIssuer, password)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Verify Keyring
		storedPass, err := service.GetSignaturePassword("1790012345001")
		assert.NoError(t, err)
		assert.Equal(t, password, storedPass)
	})

	t.Run("Update Existing Issuer Config", func(t *testing.T) {
		// Arrange
		existingIssuer := &domain.Issuer{
			BaseEntity: domain.BaseEntity{ID: 1},
			RUC:        "1790012345001",
		}
		
		updatedIssuer := &domain.Issuer{
			RUC:          "1790012345001",
			BusinessName: "Empresa Renombrada",
		}
		
		newPassword := "NewPass456"

		// Expects
		mockRepo.On("GetActive", ctx).Return(existingIssuer, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(i *domain.Issuer) bool {
			return i.ID == 1 && i.BusinessName == "Empresa Renombrada"
		})).Return(nil).Once()

		// New logic expectations
		mockEpRepo.On("GetByPoint", ctx, mock.Anything, mock.Anything, mock.Anything, "01").Return(nil, nil).Once()
		mockEpRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
		mockEpRepo.On("GetByPoint", ctx, mock.Anything, mock.Anything, mock.Anything, "04").Return(nil, nil).Once()
		mockEpRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		// Act
		err := service.SaveIssuerConfig(ctx, updatedIssuer, newPassword)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Verify Keyring Updated
		storedPass, err := service.GetSignaturePassword("1790012345001")
		assert.NoError(t, err)
		assert.Equal(t, newPassword, storedPass)
	})
}
