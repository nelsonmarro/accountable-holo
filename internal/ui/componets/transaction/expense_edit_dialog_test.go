package transaction

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

// MockCategoryService for testing
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) GetCategoryByTypeAndName(ctx context.Context, t domain.CategoryType, n string) (*domain.Category, error) {
	args := m.Called(ctx, t, n)
	return args.Get(0).(*domain.Category), args.Error(1)
}
func (m *MockCategoryService) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Category), args.Error(1)
}
func (m *MockCategoryService) GetPaginatedCategories(ctx context.Context, p, ps int, f ...string) (*domain.PaginatedResult[domain.Category], error) {
	return nil, nil
}
func (m *MockCategoryService) GetSelectablePaginatedCategories(ctx context.Context, p, ps int, f ...string) (*domain.PaginatedResult[domain.Category], error) {
	return nil, nil
}
func (m *MockCategoryService) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*domain.Category), args.Error(1)
}
func (m *MockCategoryService) CreateCategory(ctx context.Context, c *domain.Category) error { return nil }
func (m *MockCategoryService) UpdateCategory(ctx context.Context, c *domain.Category) error { return nil }
func (m *MockCategoryService) DeleteCategory(ctx context.Context, id int) error { return nil }

func TestEditExpenseDialog_RecurrenceDetection(t *testing.T) {
	app := test.NewApp()
	win := app.NewWindow("Test Edit Recurrence")

	t.Run("Should detect recurrence with exact description", func(t *testing.T) {
		mockTxSvc := new(MockTransactionService)
		mockRecurSvc := new(MockRecurringTransactionService)
		mockCatSvc := new(MockCategoryService)

		tx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: 1},
			Description: "Internet Bill",
			Amount: 50.0,
			AccountID: 1,
			CategoryID: 10,
		}

		rule := domain.RecurringTransaction{
			BaseEntity: domain.BaseEntity{ID: 5},
			Description: "Internet Bill",
			AccountID: 1,
			CategoryID: 10,
			IsActive: true,
		}

		mockTxSvc.On("GetTransactionByID", mock.Anything, mock.Anything).Return(tx, nil)
		mockCatSvc.On("GetCategoryByID", mock.Anything, mock.Anything).Return(&domain.Category{Name: "Services"}, nil)
		mockRecurSvc.On("GetAll", mock.Anything).Return([]domain.RecurringTransaction{rule}, nil)

		dlg := NewEditExpenseDialog(win, nil, mockTxSvc, mockRecurSvc, mockCatSvc, func(){}, 1, 1, domain.User{})
		dlg.Show()
		
		time.Sleep(500 * time.Millisecond)

		assert.True(t, dlg.isRecurringCheck.Checked, "Checkbox should be checked for exact match")
	})

	t.Run("Should detect recurrence with (Recurrente) suffix", func(t *testing.T) {
		mockTxSvc := new(MockTransactionService)
		mockRecurSvc := new(MockRecurringTransactionService)
		mockCatSvc := new(MockCategoryService)

		tx := &domain.Transaction{
			BaseEntity: domain.BaseEntity{ID: 2},
			Description: "Office Rent (Recurrente)",
			AccountID: 1,
			CategoryID: 20,
		}

		rule := domain.RecurringTransaction{
			Description: "Office Rent",
			AccountID: 1,
			CategoryID: 20,
			IsActive: true,
		}

		mockTxSvc.On("GetTransactionByID", mock.Anything, mock.Anything).Return(tx, nil)
		mockCatSvc.On("GetCategoryByID", mock.Anything, mock.Anything).Return(&domain.Category{Name: "Rent"}, nil)
		mockRecurSvc.On("GetAll", mock.Anything).Return([]domain.RecurringTransaction{rule}, nil)

		dlg := NewEditExpenseDialog(win, nil, mockTxSvc, mockRecurSvc, mockCatSvc, func(){}, 2, 1, domain.User{})
		dlg.Show()
		
		time.Sleep(500 * time.Millisecond)

		assert.True(t, dlg.isRecurringCheck.Checked, "Checkbox should be checked for suffixed description")
	})
}
