package ui

import (
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"

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

// setupUITestForTabs creates a test app and our UI struct with a mock service.
func setupUITestForTabs() (*UI, *mocks.MockAccountService) {
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

func setupUITest() (*UI, *mocks.MockAccountService) {
	mockService := new(mocks.MockAccountService)

	ui := NewUI(mockService)

	a := test.NewApp()
	ui.Init(a)

	// Replace the real loggers with silent ones for clean test output
	ui.infoLogger = log.New(io.Discard, "", 0)
	ui.errorLogger = log.New(io.Discard, "", 0)

	return ui, mockService
}

func waitTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		// Completed normally.
	case <-time.After(timeout):
		// Timed out.
		t.Fatal("Test timed out waiting for goroutine to finish")
	}
}
