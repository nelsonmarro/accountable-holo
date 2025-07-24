package ui

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type AccountService interface {
	GetAllAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccountByID(ctx context.Context, id int) (*domain.Account, error)
	CreateNewAccount(ctx context.Context, acc *domain.Account) error
	UpdateAccount(ctx context.Context, acc *domain.Account) error
	DeleteAccount(ctx context.Context, id int) error
}

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
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type TransactionService interface {
	CreateTransaction(ctx context.Context, transaction *domain.Transaction) error
	GetTransactionsByAccountPaginated(
		ctx context.Context,
		accountID,
		page,
		pageSize int,
		filter ...string,
	) (*domain.PaginatedResult[domain.Transaction], error)

	FindTransactionsByAccount(
		ctx context.Context,
		accountID int,
		page int,
		pageSize int,
		filters domain.TransactionFilters,
	) (*domain.PaginatedResult[domain.Transaction], error)

	GetTransactionsByDateRange(
		ctx context.Context,
		accountID int,
		startDate, endDate time.Time,
	) ([]domain.Transaction, error)

	FindAllTransactionsByAccount(
		ctx context.Context,
		accountID int,
		filters domain.TransactionFilters,
	) ([]domain.Transaction, error)

	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	VoidTransaction(ctx context.Context, transactionID int) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
	UpdateAttachmentPath(ctx context.Context, transactionID int, attachmentPath string) error
}

type StorageService interface {
	Save(ctx context.Context, sourcePath string, destinationName string) (string, error)
	GetFullPath(storagePath string) (string, error)
	Delete(ctx context.Context, storagePath string) error
}

type ReportService interface {
	GetFinancialSummary(ctx context.Context, startDate, endDate time.Time, accountID *int) (domain.FinancialSummary, error)
	GetReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time) (*domain.Reconciliation, error)
}
