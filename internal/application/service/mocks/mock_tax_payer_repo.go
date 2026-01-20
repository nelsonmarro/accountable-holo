package mocks

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockTaxPayerRepository struct {
	mock.Mock
}

func (m *MockTaxPayerRepository) GetByID(ctx context.Context, id int) (*domain.TaxPayer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxPayer), args.Error(1)
}

func (m *MockTaxPayerRepository) GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error) {
	args := m.Called(ctx, identification)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaxPayer), args.Error(1)
}

func (m *MockTaxPayerRepository) Create(ctx context.Context, tp *domain.TaxPayer) error {
	args := m.Called(ctx, tp)
	return args.Error(0)
}

func (m *MockTaxPayerRepository) Update(ctx context.Context, tp *domain.TaxPayer) error {
	args := m.Called(ctx, tp)
	return args.Error(0)
}

func (m *MockTaxPayerRepository) GetAll(ctx context.Context) ([]domain.TaxPayer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaxPayer), args.Error(1)
}

func (m *MockTaxPayerRepository) GetPaginated(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResult[domain.TaxPayer], error) {
	args := m.Called(ctx, page, pageSize, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PaginatedResult[domain.TaxPayer]), args.Error(1)
}
