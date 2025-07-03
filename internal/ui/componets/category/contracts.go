package category

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type CategoryService interface {
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}
