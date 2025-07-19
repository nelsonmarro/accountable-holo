package service

import (
	"context"

	"fyne.io/fyne/v2"
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
	GetTransactionsByAccountPaginated(
		ctx context.Context,
		accountID,
		page,
		pageSize int,
		filter ...string,
	) (*domain.PaginatedResult[domain.Transaction], error)
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	VoidTransaction(ctx context.Context, transactionID int) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
	UpdateAttachmentPath(ctx context.Context, transactionID int, attachmentPath string) error
}

type StorageService interface {
	Save(ctx context.Context, source fyne.URI, destinationName string) (string, error)
	Delete(ctx context.Context, storageURI string) error
}