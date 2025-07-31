package transaction

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// TransactionService defines the interface for transaction-related business logic.
type TransactionService interface {
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	CreateTransaction(ctx context.Context, tx *domain.Transaction) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
	VoidTransaction(ctx context.Context, id int) error
}

// CategoryService defines the interface for category-related business logic.
type CategoryService interface {
	GetCategoryByTypeAndName(ctx context.Context, catType domain.CategoryType, name string) (*domain.Category, error)
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
	GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
		*domain.PaginatedResult[domain.Category],
		error,
	)
	GetSelectablePaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
		*domain.PaginatedResult[domain.Category],
		error,
	)
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type StorageService interface {
	Save(ctx context.Context, sourcePath string, destinationName string) (string, error)
	GetFullPath(storagePath string) (string, error)
	Delete(ctx context.Context, storagePath string) error
}
