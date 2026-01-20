package mocks

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRecurringTransactionRepository struct {
	mock.Mock
}

func (m *MockRecurringTransactionRepository) Create(ctx context.Context, rt *domain.RecurringTransaction) error {
	args := m.Called(ctx, rt)
	return args.Error(0)
}

func (m *MockRecurringTransactionRepository) GetAll(ctx context.Context) ([]domain.RecurringTransaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringTransactionRepository) GetAllActive(ctx context.Context) ([]domain.RecurringTransaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringTransactionRepository) Update(ctx context.Context, rt *domain.RecurringTransaction) error {
	args := m.Called(ctx, rt)
	return args.Error(0)
}

func (m *MockRecurringTransactionRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
