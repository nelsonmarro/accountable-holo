package ui

import (
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/mock"
)

func TestRefreshAccountList(t *testing.T) {
	t.Run("should populate accounts slice on successful service call", func(t *testing.T) {
		// Arrange
		ui, mockService := setupUITest()

		// This is the data we expect the service to return
		sampleAccounts := []domain.Account{
			{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Account 1"},
			{BaseEntity: domain.BaseEntity{ID: 2}, Name: "Account 2"},
		}

		mockService.On("GetAllAccounts", mock.Anything).Return(sampleAccounts, nil)
	})
}
