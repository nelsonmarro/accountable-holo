package ui

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

type AccountService interface {
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
	GetSelectablePaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
		*domain.PaginatedResult[domain.Category],
		error,
	)
	GetCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type TransactionService interface {
	GetTransactionByAccountPaginated(ctx context.Context, accountID, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Transaction], error)
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	VoidTransaction(ctx context.Context, transactionID int) error
	CreateTransaction(ctx context.Context, tx *domain.Transaction) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
}

type ReportService interface {
	GenerateFinancialSummary(ctx context.Context, startDate, endDate time.Time, accountID *int) (domain.FinancialSummary, error)
	GenerateReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time, endingBalance decimal.Decimal) (*domain.Reconciliation, error)
}
