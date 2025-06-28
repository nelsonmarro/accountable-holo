package account

import (
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/nelsonmarro/accountable-holo/internal/ui/mocks"
)

// TestMain is the entry point for all tests in this package.
func TestMain(m *testing.M) {
	// --- Run the tests ---
	code := m.Run()
	os.Exit(code)
}

func setupDependencies() (fyne.Window, *log.Logger, *mocks.MockAccountService) {
	// Create a Fyne app and window that run only in memory.
	test.NewApp()
	win := test.NewWindow(nil)

	// Create our mock service.
	mockService := new(mocks.MockAccountService)

	// Use a silent logger that discards output.
	silentLogger := log.New(io.Discard, "", 0)

	return win, silentLogger, mockService
}

func setupTestAddDialog(callback func()) (*AddAccountDialog, *mocks.MockAccountService) {
	win, silentLogger, mockService := setupDependencies()

	// Create the dialog handler instance with all our test objects.
	dialogHandler := NewAddAccountDialog(win, silentLogger, mockService, callback)

	// Pre-populate the entry widgets with some valid test data.
	test.Type(dialogHandler.nameEntry, "Test Bank Account")
	test.Type(dialogHandler.nameEntry, "Test Bank Account")
	test.Type(dialogHandler.tipoSelect, "Ahorros")
	test.Type(dialogHandler.amountEntry, "150.75")
	test.Type(dialogHandler.numberEntry, "123456789")

	return dialogHandler, mockService
}

func setupTestEditDialog(callback func()) (*AddAccountDialog, *mocks.MockAccountService) {
	win, silentLogger, mockService := setupDependencies()

	// Create the dialog handler instance with all our test objects.
	dialogHandler := NewAddAccountDialog(win, silentLogger, mockService, callback)

	// Pre-populate the entry widgets with some valid test data.
	test.Type(dialogHandler.nameEntry, "Test Bank Account")
	test.Type(dialogHandler.nameEntry, "Test Bank Account")
	test.Type(dialogHandler.tipoSelect, "Ahorros")
	test.Type(dialogHandler.amountEntry, "150.75")
	test.Type(dialogHandler.numberEntry, "123456789")

	return dialogHandler, mockService
}

func setupTestDelDialog(callback func()) (*DeleteAccountDialog, *mocks.MockAccountService) {
	win, silentLogger, mockService := setupDependencies()

	// Create the dialog handler instance with all our test objects.
	dialogHandler := NewDeleteAccountDialog(win, silentLogger, mockService, callback, 1)

	return dialogHandler, mockService
}

// waitTimeout is a helper to wait for a WaitGroup with a safety timeout.
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
