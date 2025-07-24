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

		ui, mockService := setupUITestForTabs()

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
		ui.loadAccounts()

		waitTimeout(t, &wg, 1*time.Second)

		mockService.AssertExpectations(t)
		assert.Equal(t, sampleAccounts, ui.accounts)
	})

	t.Run("should not change accounts slice when service returns an error", func(t *testing.T) {
		// Arrange
		var wg sync.WaitGroup
		wg.Add(1)

		ui, mockService := setupUITestForTabs()

		// Pre-populate with some initial data
		ui.accounts = []domain.Account{
			{BaseEntity: domain.BaseEntity{ID: 99}, Name: "Existing Account"},
		}

		// Configure the mock to return an error
		mockService.On("GetAllAccounts", mock.Anything).
			Return([]domain.Account{}, errors.New("database is down")).
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		ui.accountList = widget.NewList(
			func() int { return 1 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i widget.ListItemID, o fyne.CanvasObject) {},
		)

		// Act
		ui.loadAccounts()

		// Assert
		waitTimeout(t, &wg, 1*time.Second)

		mockService.AssertExpectations(t)
		// The key assertion: the original data was NOT overwritten on error.
		require.Len(t, ui.accounts, 1)
		assert.Equal(t, 99, ui.accounts[0].ID)
	})
}

func TestFillListData(t *testing.T) {
	// Arrange
	ui, _ := setupUITestForTabs()

	// Manually populate the ui.accounts slice with data for the test.
	ui.accounts = []domain.Account{
		{
			BaseEntity:     domain.BaseEntity{ID: 101},
			Number:         "22",
			Name:           "Test Checking",
			Type:           domain.OrdinaryAccount,
			InitialBalance: 1234.56,
		},
		{
			BaseEntity:     domain.BaseEntity{ID: 102},
			Number:         "23",
			Name:           "Test Savings",
			Type:           domain.SavingAccount,
			InitialBalance: 789.00,
		},
	}

	// Create a template canvas object, just like Fyne would.
	listItemUI := ui.makeAccountListUI()

	// Act
	// Simulate Fyne calling this function for the first item in the list (ID 101).
	ui.fillAccountListData(0, listItemUI)

	// Assert
	// We need to "drill down" into the container hierarchy to find the widgets.
	borderContainer := listItemUI.(*fyne.Container)

	// Assert Labels are set correctly
	nameLbl := borderContainer.Objects[0].(*widget.Label)
	assert.Equal(t, "Test Checking - 22", nameLbl.Text)

	typeLbl := borderContainer.Objects[2].(*widget.Label)
	assert.Equal(t, "Tipo de Cuenta: Corriente", typeLbl.Text)

	balanceLbl := borderContainer.Objects[3].(*widget.Label)
	assert.Equal(t, "1234.56", balanceLbl.Text)

	// Assert that the buttons have an OnTapped handler assigned.
	actionsContainer := borderContainer.Objects[4].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	deleteBtn := actionsContainer.Objects[1].(*widget.Button)

	assert.NotNil(t, editBtn.OnTapped, "Edit button should have an OnTapped handler")
	assert.NotNil(t, deleteBtn.OnTapped, "Delete button should have an OnTapped handler")
}
