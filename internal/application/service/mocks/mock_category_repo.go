package mocks

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) FindByNameAndType(ctx context.Context, name string, catType domain.CategoryType) (*domain.Category, error) {
	args := m.Called(ctx, name, catType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Category], error) {
	args := m.Called(ctx, page, pageSize, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PaginatedResult[domain.Category]), args.Error(1)
}

func (m *MockCategoryRepository) GetSelectablePaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Category], error) {
	args := m.Called(ctx, page, pageSize, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PaginatedResult[domain.Category]), args.Error(1)
}

func (m *MockCategoryRepository) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) CategoryExists(ctx context.Context, name string, id int) (bool, error) {
	args := m.Called(ctx, name, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) CreateCategory(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) UpdateCategory(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
