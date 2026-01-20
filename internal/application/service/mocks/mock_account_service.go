package mocks

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) GetAllAccounts(ctx context.Context) ([]domain.Account, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Account), args.Error(1)
}

func (m *MockAccountService) GetAccountByID(ctx context.Context, id int) (*domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountService) CreateNewAccount(ctx context.Context, acc *domain.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockAccountService) UpdateAccount(ctx context.Context, acc *domain.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockAccountService) DeleteAccount(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
