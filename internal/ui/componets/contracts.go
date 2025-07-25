package componets

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type CategoryService interface {
	GetPaginatedCategories(
		ctx context.Context,
		page,
		pageSize int,
		filter ...string,
	) (*domain.PaginatedResult[domain.Category], error)

	GetSelectablePaginatedCategories(
		ctx context.Context,
		page, pageSize int,
		filter ...string,
	) (*domain.PaginatedResult[domain.Category], error)

	GetAllCategories(ctx context.Context) ([]domain.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
}

