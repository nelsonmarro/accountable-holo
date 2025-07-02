// Package mocks is a package that provides mock implementations of the domain interfaces for testing purposes.
package mocks

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/mock"
)

// ---- MockAccountRepository ----

// MockAccountRepository is a mock type for the AccountRepository interface
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) CreateAccount(ctx context.Context, acc *domain.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockAccountRepository) GetAllAccounts(ctx context.Context) ([]domain.Account, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Account), args.Error(1)
}

func (m *MockAccountRepository) GetAccountByID(ctx context.Context, id int) (*domain.Account, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) UpdateAccount(ctx context.Context, acc *domain.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockAccountRepository) DeleteAccount(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccountRepository) AccountExists(ctx context.Context, name, number string, id int) (bool, error) {
	args := m.Called(ctx, name, number, id)
	return args.Bool(0), args.Error(1)
}

// ---- End of MockAccountRepository ----
