package transaction

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

// MockTransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateTransaction(ctx context.Context, tx *domain.Transaction, user domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

// ... other stubs ...

type MockRecurringTransactionService struct {
	mock.Mock
}

func (m *MockRecurringTransactionService) Create(ctx context.Context, rt *domain.RecurringTransaction) error {
	return nil
}
func (m *MockRecurringTransactionService) GetAll(ctx context.Context) ([]domain.RecurringTransaction, error) {
	return nil, nil
}
func (m *MockRecurringTransactionService) Update(ctx context.Context, rt *domain.RecurringTransaction) error {
	return nil
}
func (m *MockRecurringTransactionService) Delete(ctx context.Context, id int) error {
	return nil
}
func (m *MockRecurringTransactionService) ProcessPendingRecurrences(ctx context.Context, user domain.User) error {
	return nil
}

// Implement other interface methods with empty stubs to satisfy the interface
func (m *MockTransactionService) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	return nil, nil
}
func (m *MockTransactionService) GetItemsByTransactionID(ctx context.Context, id int) ([]domain.TransactionItem, error) {
	return nil, nil
}
func (m *MockTransactionService) UpdateTransaction(ctx context.Context, tx *domain.Transaction, user domain.User) error {
	return nil
}
func (m *MockTransactionService) VoidTransaction(ctx context.Context, id int, user domain.User) (int, error) {
	return 0, nil
}
func (m *MockTransactionService) RevertVoidTransaction(ctx context.Context, id int) error {
	return nil
}
func (m *MockTransactionService) ReconcileAccount(ctx context.Context, accID int, start, end time.Time, bal decimal.Decimal) (*domain.Reconciliation, error) {
	return nil, nil
}
func (m *MockTransactionService) FindTransactionsByAccount(ctx context.Context, accountID, page, pageSize int, filters domain.TransactionFilters, search *string) (*domain.PaginatedResult[domain.Transaction], error) {
	return nil, nil
}
func (m *MockTransactionService) FindAllTransactionsByAccount(ctx context.Context, accountID int, filters domain.TransactionFilters) ([]domain.Transaction, error) {
	return nil, nil
}

// TestAddExpenseDialog_Submit
func TestAddExpenseDialog_Submit(t *testing.T) {
	app := test.NewApp()
	win := app.NewWindow("Test Expense")

	mockTxService := new(MockTransactionService)
	// We don't need real mocks for others if we don't trigger them or pass nil/empty
	// But AddExpenseDialog expects interfaces.
	
	// Create dialog instance
	// Note: We need minimal dependencies.
	dlg := NewAddExpenseDialog(
		win,
		nil, // logger
		mockTxService,
		nil, // RecurService (not testing recurrence here)
		nil, // CategoryService
		func() {}, // callback
		1, // accountID
		domain.User{},
	)

	// Case 1: Validation Fail (Empty Amount)
	dlg.amountEntry.SetText("")
	dlg.submit() // Should show error dialog (mocked by Fyne test app but won't crash)
	mockTxService.AssertNotCalled(t, "CreateTransaction")

	// Case 2: Validation Fail (No Category)
	dlg.amountEntry.SetText("100.50")
	dlg.submit()
	mockTxService.AssertNotCalled(t, "CreateTransaction")

	// Case 3: Success
	// Manually inject category since we can't open search dialog
	cat := &domain.Category{BaseEntity: domain.BaseEntity{ID: 10}, Name: "Office Supplies", Type: domain.Outcome}
	// AddExpenseDialog struct fields are private, but in same package 'transaction' we can access them!
	dlg.selectedCategory = cat 
	dlg.descriptionEntry.SetText("Paper and Pens")
	
	// Setup expectation
	mockTxService.On("CreateTransaction", mock.Anything, mock.MatchedBy(func(tx *domain.Transaction) bool {
		return tx.Amount == 100.50 && 
		       tx.Description == "Paper and Pens" &&
		       tx.CategoryID == 10 &&
		       len(tx.Items) == 1 &&
		       tx.Items[0].UnitPrice == 100.50
	}), mock.Anything).Return(nil)

	dlg.submit()

	mockTxService.AssertExpectations(t)
}
