package ui

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewUI(t *testing.T) {
	// Arrange
	mockService := new(mocks.MockAccountService)

	// Act
	ui := NewUI(mockService)

	// Assert
	assert.NotNil(t, ui, "UI object should not be nil")
	assert.NotNil(t, ui.app, "Fyne app should be initialized")
	assert.NotNil(t, ui.mainWindow, "Main window should be initialized")
	assert.Equal(t, "Accountable Holo", ui.mainWindow.Title())
	assert.Equal(t, mockService, ui.accService, "Account service dependency should be set")

	// Check if our custom theme was applied
	_, ok := ui.app.Settings().Theme().(*AppTheme)
	assert.True(t, ok, "Expected custom AppTheme to be set")
}

func TestBuildMainUI(t *testing.T) {
	// Arrange
	ui, mockService := setupUITest()
	var wg sync.WaitGroup
	wg.Add(1)

	// The makeAccountTab function calls refreshAccountList, which calls this service method.
	mockService.On("GetAllAccounts", mock.Anything).
		Return([]domain.Account{}, nil).
		Run(func(args mock.Arguments) {
			defer wg.Done()
		})

	// Act
	ui.buildMainUI()

	// Assert
	waitTimeout(t, &wg, 2*time.Second)

	// 1. Verify the window content was set and is the correct type.
	content := ui.mainWindow.Content()
	require.NotNil(t, content, "Window content should be set")

	// 2. Verify the content is a tab container.
	tabs, ok := content.(*container.AppTabs)
	require.True(t, ok, "Window content should be an AppTabs container")

	// 3. Verify the tabs were created correctly.
	assert.Len(t, tabs.Items, 3, "Should be exactly 3 tabs")
	assert.Equal(t, "Accounts", tabs.Items[0].Text)
	assert.Equal(t, "Transactions", tabs.Items[1].Text)
	assert.Equal(t, "Reports", tabs.Items[2].Text)

	// 4. Verify the service method was called.
	mockService.AssertExpectations(t)
}
