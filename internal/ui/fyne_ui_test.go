package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/nelsonmarro/accountable-holo/internal/ui/mocks"
	"github.com/stretchr/testify/assert"
)

// TestNewUI now tests the simple constructor.
func TestNewUI(t *testing.T) {
	// Arrange
	mockService := new(mocks.MockAccountService)
	services := &Services{
		AccService: mockService,
	}

	// Act
	ui := NewUI(services)

	// Assert
	assert.NotNil(t, ui, "UI object should not be nil")
	assert.Nil(t, ui.app, "Fyne app should be nil before Init is called")
	assert.Nil(t, ui.mainWindow, "Main window should be nil before Init is called")
	assert.Equal(t, mockService, ui.Services.AccService, "Account service dependency should be set")
	assert.NotNil(t, ui.infoLogger)
	assert.NotNil(t, ui.errorLogger)
}

// TestUI_Init tests the Fyne-specific initialization.
func TestUI_Init(t *testing.T) {
	// Arrange
	testApp := test.NewApp()
	mockService := new(mocks.MockAccountService)
	services := &Services{
		AccService: mockService,
	}
	ui := NewUI(services)

	// Act
	ui.Init(testApp)

	// Assert
	assert.Equal(t, testApp, ui.app, "Fyne app should be set")
	assert.NotNil(t, ui.mainWindow, "Main window should be initialized")
	assert.Equal(t, "Accountable Holo", ui.mainWindow.Title())

	// Check if our custom theme was applied correctly
	_, ok := ui.app.Settings().Theme().(*AppTheme)
	assert.True(t, ok, "Expected custom AppTheme to be set")
}

// TODO: Refator TestBuildMainUI
// func TestBuildMainUI(t *testing.T) {
// 	// Arrange
// 	ui, mockService := setupUITest()
// 	var wg sync.WaitGroup
// 	wg.Add(1)
//
// 	// The makeAccountTab function calls refreshAccountList, which calls this service method.
// 	mockService.On("GetAllAccounts", mock.Anything).
// 		Return([]domain.Account{}, nil).
// 		Run(func(args mock.Arguments) {
// 			defer wg.Done()
// 		})
//
// 	// Act
// 	ui.buildMainUI()
//
// 	// Assert
// 	waitTimeout(t, &wg, 2*time.Second)
//
// 	// 1. Verify the window content was set and is the correct type.
// 	content := ui.mainWindow.Content()
// 	require.NotNil(t, content, "Window content should be set")
//
// 	// 2. Verify the content is a tab container.
// 	tabs, ok := content.(*container.AppTabs)
// 	require.True(t, ok, "Window content should be an AppTabs container")
//
// 	// 3. Verify the tabs were created correctly.
// 	assert.Len(t, tabs.Items, 3, "Should be exactly 3 tabs")
// 	assert.Equal(t, "Cuentas", tabs.Items[0].Text)
// 	assert.Equal(t, "Finanzas", tabs.Items[1].Text)
// 	assert.Equal(t, "Reportes", tabs.Items[2].Text)
//
// 	// 4. Verify the service method was called.
// 	mockService.AssertExpectations(t)
// }
