package service

import (
	"context"
	"time"

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
	CategoryExists(ctx context.Context, name string, id int) (bool, error)
	CreateCategory(ctx context.Context, category *domain.Category) error
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *domain.Transaction) error
	FindTransactionsByAccount(
		ctx context.Context,
		accountID int,
		page int,
		pageSize int,
		filters domain.TransactionFilters,
		searchString *string,
	) (*domain.PaginatedResult[domain.Transaction], error)

	FindAllTransactionsByAccount(
		ctx context.Context,
		accountID int,
		filters domain.TransactionFilters,
		searchString *string,
	) ([]domain.Transaction, error)

	FindAllTransactions(
		ctx context.Context,
		filters domain.TransactionFilters,
		searchString *string,
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

type ReportRepository interface {
	GetFinancialSummary(ctx context.Context, startDate, endDate time.Time, accountID *int) (domain.FinancialSummary, error)
	GetReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time) (*domain.Reconciliation, error)
}

type ReportGenerator interface {
	SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string) error
}
