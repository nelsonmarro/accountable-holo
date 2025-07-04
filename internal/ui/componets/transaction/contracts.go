package transaction

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// TransactionService defines the interface for transaction-related business logic.
type TransactionService interface {
	GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error)
	CreateNewTransaction(ctx context.Context, tx *domain.Transaction) error
	UpdateTransaction(ctx context.Context, tx *domain.Transaction) error
	VoidTransaction(ctx context.Context, id int) error
}

// CategoryService defines the interface for category-related business logic.
type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
}
