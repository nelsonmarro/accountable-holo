package transaction

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, tx *domain.Transaction) error
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
	VoidTransaction(ctx context.Context, transactionID int) error
}

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
	GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Category], error)
	GetSelectablePaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Category], error)
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}