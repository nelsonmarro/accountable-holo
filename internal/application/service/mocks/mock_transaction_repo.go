package mocks

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindTransactionsByAccount(
	ctx context.Context,
	accountID int,
	page int,
	pageSize int,
	filters domain.TransactionFilters,
	searchString *string,
) (*domain.PaginatedResult[domain.Transaction], error) {
	args := m.Called(ctx, accountID, page, pageSize, filters, searchString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PaginatedResult[domain.Transaction]), args.Error(1)
}

func (m *MockTransactionRepository) FindAllTransactionsByAccount(
	ctx context.Context,
	accountID int,
	filters domain.TransactionFilters,
	searchString *string,
) ([]domain.Transaction, error) {
	args := m.Called(ctx, accountID, filters, searchString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindAllTransactions(
	ctx context.Context,
	filters domain.TransactionFilters,
	searchString *string,
) ([]domain.Transaction, error) {
	args := m.Called(ctx, filters, searchString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetBalanceAsOf(ctx context.Context, accountID int, date time.Time) (decimal.Decimal, error) {
	args := m.Called(ctx, accountID, date)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockTransactionRepository) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetItemsByTransactionID(ctx context.Context, transactionID int) ([]domain.TransactionItem, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TransactionItem), args.Error(1)
}

func (m *MockTransactionRepository) VoidTransaction(ctx context.Context, transactionID int, currentUser domain.User) (int, error) {
	args := m.Called(ctx, transactionID, currentUser)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionRepository) RevertVoidTransaction(ctx context.Context, voidTransactionID int) error {
	args := m.Called(ctx, voidTransactionID)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateTransaction(ctx context.Context, tx *domain.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateAttachmentPath(ctx context.Context, transactionID int, attachmentPath string) error {
	args := m.Called(ctx, transactionID, attachmentPath)
	return args.Error(0)
}