package mocks

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockIssuerRepository struct {
	mock.Mock
}

func (m *MockIssuerRepository) GetActive(ctx context.Context) (*domain.Issuer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Issuer), args.Error(1)
}

func (m *MockIssuerRepository) Create(ctx context.Context, issuer *domain.Issuer) error {
	args := m.Called(ctx, issuer)
	return args.Error(0)
}

func (m *MockIssuerRepository) Update(ctx context.Context, issuer *domain.Issuer) error {
	args := m.Called(ctx, issuer)
	return args.Error(0)
}