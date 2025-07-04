package ui

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type AccountService interface {
	GetPaginatedAccounts(ctx context.Context, page, pageSize int, filter string) (*domain.PaginatedResult[domain.Account], error)
	GetAllAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccountByID(ctx context.Context, id int) (*domain.Account, error)
	CreateNewAccount(ctx context.Context, acc *domain.Account) error
	UpdateAccount(ctx context.Context, acc *domain.Account) error
	DeleteAccount(ctx context.Context, id int) error
}

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
	GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
		*domain.PaginatedResult[domain.Category],
		error,
	)
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type TransactionService interface {
	CreateNewTransaction(ctx context.Context, tx *domain.Transaction) error
	GetTransactionByAccountPaginated(ctx context.Context, accountID, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Transaction], error)
	VoidTransaction(ctx context.Context, transactionID int) error
}
