package transaction

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

type RecurringTransactionService interface {
	Create(ctx context.Context, rt *domain.RecurringTransaction) error
}

// TransactionService defines the interface for transaction-related business logic.
type TransactionService interface {
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	GetItemsByTransactionID(ctx context.Context, transactionID int) ([]domain.TransactionItem, error)
	CreateTransaction(ctx context.Context, transaction *domain.Transaction, currentUser domain.User) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction, currentUser domain.User) error
	VoidTransaction(ctx context.Context, id int, currentUser domain.User) error
	ReconcileAccount(
		ctx context.Context,
		accountID int,
		startDate time.Time,
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
	GenerateReconciliationReportFile(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string, currentUser *domain.User) error
}

type SriService interface {
	EmitirFactura(ctx context.Context, transactionID int, signaturePassword string) error
	GenerateRide(ctx context.Context, transactionID int) (string, error)
	SyncReceipt(ctx context.Context, receipt *domain.ElectronicReceipt) (string, error)
	ProcessBackgroundSync(ctx context.Context) (int, error)
}

type TaxPayerService interface {
	GetByID(ctx context.Context, id int) (*domain.TaxPayer, error)
	GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error)
	Create(ctx context.Context, tp *domain.TaxPayer) error
	Update(ctx context.Context, tp *domain.TaxPayer) error
	GetPaginated(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResult[domain.TaxPayer], error)
}
