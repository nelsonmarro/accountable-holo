package ui

import (
	"io"
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/mocks"
)

// TestMain is the entry point for all tests in this package.
func TestMain(m *testing.M) {
	// --- Run the tests ---
	code := m.Run()
	os.Exit(code)
}

// setupUITest creates a test app and our UI struct with a mock service.
func setupUITest() (*UI, *mocks.MockAccountService) {
	// Create a test app and window (runs in memory)
	a := test.NewApp()
	w := test.NewWindow(nil)

	mockService := new(mocks.MockAccountService)
	silentLogger := log.New(io.Discard, "", 0)

	// Create the UI instance to be tested
	ui := &UI{
		mainWindow:  w,
		app:         a,
		errorLogger: silentLogger,
		accService:  mockService,
		accounts:    make([]domain.Account, 0), // Start with an empty slice
	}

	return ui, mockService
}
