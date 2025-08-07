package transaction

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

// TransactionService defines the interface for transaction-related business logic.
type TransactionService interface {
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	CreateTransaction(ctx context.Context, transaction *domain.Transaction, currentUser *domain.User) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction, currentUser *domain.User) error
	VoidTransaction(ctx context.Context, id int) error
	ReconcileAccount(
		ctx context.Context,
		accountID int,
		endDate time.Time,
		actualEndingBalance decimal.Decimal,
	) (*domain.Reconciliation, error)
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

type ReportService interface {
	GenerateReconciliationReportFile(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string) error
}
