package ui

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshAccountList(t *testing.T) {
	t.Run("should populate accounts slice on successful service call", func(t *testing.T) {
		// Arrange
		var wg sync.WaitGroup
		wg.Add(1)

		ui, mockService := setupUITest()

		// This is the data we expect the service to return
		sampleAccounts := []domain.Account{
			{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Account 1"},
			{BaseEntity: domain.BaseEntity{ID: 2}, Name: "Account 2"},
		}

		mockService.On("GetAllAccounts", mock.Anything).Return(sampleAccounts, nil)

		// Create a placeholder list widget for the UI struct
		ui.accountList = widget.NewList(
			func() int { return 2 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, o fyne.CanvasObject) {},
		)

		// Act
		ui.refreshAccountList()

		time.Sleep(100 * time.Millisecond)

		mockService.AssertExpectations(t)
		assert.Equal(t, sampleAccounts, ui.accounts)
	})
}
