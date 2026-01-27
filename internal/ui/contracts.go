package ui

import (
	"context"
	"time"

	"github.com/nelsonmarro/verith/internal/domain"
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
	GetCategoryByTypeAndName(ctx context.Context, catType domain.CategoryType, name string) (*domain.Category, error)
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
	CreateTransaction(ctx context.Context, transaction *domain.Transaction, currentUser domain.User) error

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
	) ([]domain.Transaction, error)

	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	GetItemsByTransactionID(ctx context.Context, transactionID int) ([]domain.TransactionItem, error)
	VoidTransaction(ctx context.Context, transactionID int, currentUser domain.User) (int, error)
	RevertVoidTransaction(ctx context.Context, voidTransactionID int) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction, currentUser domain.User) error
	ReconcileAccount(
		ctx context.Context,
		accountID int,
		startDate time.Time,
		endDate time.Time,
		actualEndingBalance decimal.Decimal,
	) (*domain.Reconciliation, error)
}

type StorageService interface {
	Save(ctx context.Context, sourcePath string, destinationName string) (string, error)
	GetFullPath(storagePath string) (string, error)
	Delete(ctx context.Context, storagePath string) error
}

type ReportService interface {
	GetFinancialSummary(ctx context.Context, startDate, endDate time.Time, accountID *int) (domain.FinancialSummary, error)
	GetBudgetOverview(ctx context.Context, startDate, endDate time.Time) ([]domain.BudgetStatus, error)
	GetReconciliation(ctx context.Context, accountID int, startDate, endDate time.Time, endingBalance decimal.Decimal) (*domain.Reconciliation, error)
	GenerateReportFile(ctx context.Context, format string, transactions []domain.Transaction, outputPath string, currentUser *domain.User) error
	GenerateReconciliationReportFile(ctx context.Context, reconciliation *domain.Reconciliation, outputPath string, currentUser *domain.User) error
	GenerateDailyReport(ctx context.Context, accountID int) (*domain.DailyReport, error)
	GenerateDailyReportFile(ctx context.Context, report *domain.DailyReport, outputPath string, format string, currentUser *domain.User) error
}

type RecurringTransactionService interface {
	Create(ctx context.Context, rt *domain.RecurringTransaction) error
	GetAll(ctx context.Context) ([]domain.RecurringTransaction, error)
	Update(ctx context.Context, rt *domain.RecurringTransaction) error
	Delete(ctx context.Context, id int) error
	ProcessPendingRecurrences(ctx context.Context, systemUser domain.User) error
}

type IssuerService interface {
	GetActive(ctx context.Context) (*domain.Issuer, error)
	SaveIssuerConfig(ctx context.Context, issuer *domain.Issuer, password string) error
	GetIssuerConfig(ctx context.Context) (*domain.Issuer, error)
	GetSignaturePassword(ruc string) (string, error)
	GetEmissionPoints(ctx context.Context) ([]domain.EmissionPoint, error)
	UpdateEmissionPoint(ctx context.Context, ep *domain.EmissionPoint) error
}

type SriService interface {
	EmitirFactura(ctx context.Context, transactionID int, signaturePassword string) error
	EmitirNotaCredito(ctx context.Context, voidTxID int, originalTxID int, motivo string, signaturePassword string) (string, error)
	GenerateRide(ctx context.Context, transactionID int) (string, error)
	SyncReceipt(ctx context.Context, receipt *domain.ElectronicReceipt) (string, error)
	ProcessBackgroundSync(ctx context.Context) (int, error)
	ResendEmail(ctx context.Context, transactionID int) error
	GetPendingQueue(ctx context.Context) ([]domain.ElectronicReceipt, error)
}

type UserService interface {
	Login(ctx context.Context, username, password string) (*domain.User, error)
	CreateUser(ctx context.Context, username, password, firstName, lastName string, role domain.UserRole, currentUser *domain.User) error
	UpdateUser(ctx context.Context, id int, username, password, firstName, lastName string, role domain.UserRole, currentUser *domain.User) error
	DeleteUser(ctx context.Context, id int, currentUser *domain.User) error
	GetAllUsers(ctx context.Context, currentUser *domain.User) ([]domain.User, error)
	HasUsers(ctx context.Context) (bool, error)
	GetAdminUsers(ctx context.Context) ([]domain.User, error)
	ResetPassword(ctx context.Context, username, newPassword string) error
}

type TaxPayerService interface {
	GetByID(ctx context.Context, id int) (*domain.TaxPayer, error)
	GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error)
	Create(ctx context.Context, tp *domain.TaxPayer) error
	Update(ctx context.Context, tp *domain.TaxPayer) error
	Search(ctx context.Context, query string) ([]domain.TaxPayer, error)
	GetPaginated(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResult[domain.TaxPayer], error)
}
