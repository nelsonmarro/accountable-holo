package ui

import (
	"errors"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

		mockService.On("GetAllAccounts", mock.Anything).Return(sampleAccounts, nil).
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		// Create a placeholder list widget for the UI struct
		ui.accountList = widget.NewList(
			func() int { return 2 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, o fyne.CanvasObject) {},
		)

		// Act
		ui.refreshAccountList()

		waitTimeout(t, &wg, 1*time.Second)

		mockService.AssertExpectations(t)
		assert.Equal(t, sampleAccounts, ui.accounts)
	})

	t.Run("should not change accounts slice when service returns an error", func(t *testing.T) {
		// Arrange
		var wg sync.WaitGroup
		wg.Add(1)

		ui, mockService := setupUITest()

		// Pre-populate with some initial data
		ui.accounts = []domain.Account{
			{BaseEntity: domain.BaseEntity{ID: 99}, Name: "Existing Account"},
		}

		// Configure the mock to return an error
		mockService.On("GetAllAccounts", mock.Anything).
			Return(nil, errors.New("database is down")).
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		ui.accountList = widget.NewList(
			func() int { return 1 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, o fyne.CanvasObject) {},
		)

		// Act
		ui.refreshAccountList()

		// Assert
		waitTimeout(t, &wg, 1*time.Second)

		mockService.AssertExpectations(t)
		// The key assertion: the original data was NOT overwritten on error.
		require.Len(t, ui.accounts, 1)
		assert.Equal(t, 99, ui.accounts[0].ID)
	})
}
