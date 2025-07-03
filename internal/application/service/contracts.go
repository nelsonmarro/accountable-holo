package service

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type AccountRepository interface {
	GetAllAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccountByID(ctx context.Context, id int) (*domain.Account, error)
	AccountExists(ctx context.Context, name, number string, id int) (bool, error)
	CreateAccount(ctx context.Context, acc *domain.Account) error
	UpdateAccount(ctx context.Context, acc *domain.Account) error
	DeleteAccount(ctx context.Context, id int) error
}

type CategoryRepository interface {
	GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Category], error)
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	CategoryExists(ctx context.Context, name string, id int) (bool, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *domain.Transaction) error
	GetAllTransactions(ctx context.Context) ([]domain.Transaction, error)
}
