package service

import (
	"context"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
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
	FindByNameAndType(ctx context.Context, name string, catType domain.CategoryType) (*domain.Category, error)
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

	GetBalanceAsOf(ctx context.Context, accountID int, date time.Time) (decimal.Decimal, error)
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	GetItemsByTransactionID(ctx context.Context, transactionID int) ([]domain.TransactionItem, error)
	VoidTransaction(ctx context.Context, transactionID int, currentUser domain.User) (int, error)
	RevertVoidTransaction(ctx context.Context, voidTransactionID int) error
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

type RecurringTransactionRepository interface {
	Create(ctx context.Context, rt *domain.RecurringTransaction) error
	GetAll(ctx context.Context) ([]domain.RecurringTransaction, error)
	GetAllActive(ctx context.Context) ([]domain.RecurringTransaction, error)
	Update(ctx context.Context, rt *domain.RecurringTransaction) error
	Delete(ctx context.Context, id int) error
}

type IssuerRepository interface {
	GetActive(ctx context.Context) (*domain.Issuer, error)
	Create(ctx context.Context, issuer *domain.Issuer) error
	Update(ctx context.Context, issuer *domain.Issuer) error
}

type TaxPayerRepository interface {
	GetByID(ctx context.Context, id int) (*domain.TaxPayer, error)
	GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error)
	Create(ctx context.Context, tp *domain.TaxPayer) error
	Update(ctx context.Context, tp *domain.TaxPayer) error
	GetAll(ctx context.Context) ([]domain.TaxPayer, error)
	GetPaginated(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResult[domain.TaxPayer], error)
}

type EmissionPointRepository interface {
	GetByPoint(ctx context.Context, issuerID int, estCode, pointCode, receiptType string) (*domain.EmissionPoint, error)
	GetAllByIssuer(ctx context.Context, issuerID int) ([]domain.EmissionPoint, error)
	IncrementSequence(ctx context.Context, id int) error
	Create(ctx context.Context, ep *domain.EmissionPoint) error
	Update(ctx context.Context, ep *domain.EmissionPoint) error
}

type ElectronicReceiptRepository interface {
	Create(ctx context.Context, er *domain.ElectronicReceipt) error
	Update(ctx context.Context, er *domain.ElectronicReceipt) error
	UpdateStatus(ctx context.Context, accessKey string, status string, message string, authDate *time.Time) error
	UpdateXML(ctx context.Context, accessKey string, xmlContent string) error
	UpdateTaxPayerID(ctx context.Context, accessKey string, taxPayerID int) error
	UpdateEmailSent(ctx context.Context, accessKey string, sent bool) error
	GetByAccessKey(ctx context.Context, accessKey string) (*domain.ElectronicReceipt, error)
	FindPendingReceipts(ctx context.Context) ([]domain.ElectronicReceipt, error)
}
