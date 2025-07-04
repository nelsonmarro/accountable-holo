package service

import (
	"context"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TransactionServiceImpl struct {
	repo TransactionRepository
}

func NewTransactionService(repo TransactionRepository) *TransactionServiceImpl {
	return &TransactionServiceImpl{repo: repo}
}

func (s *TransactionServiceImpl) GetTransactionByAccountPaginated(ctx context.Context, accountID, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Transaction], error) {
	if accountID <= 0 {
		return nil, fmt.Errorf("ID de cuenta inválido: %d", accountID)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}
	if page > 100 {
		page = 100 // Limit to 100 pages
	}
	return s.repo.GetTransactionsByAccountPaginated(ctx, accountID, page, pageSize, filter...)
}

func (s *TransactionServiceImpl) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transacción no puede ser nula")
	}
	txValidator := validator.New().For(tx)
	txValidator.Required("Amount", "Description", "TransactionDate", "AccountID", "CategoryID")
	txValidator.NumberMin(0, "Amount")

	err := txValidator.ConsolidateErrors()
	if err != nil {
		return err
	}

	err = s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("error al crear la transacción: %w", err)
	}

	return nil
}

func (s *TransactionServiceImpl) VoidTransaction(ctx context.Context, transactionID int) error {
	if transactionID <= 0 {
		return fmt.Errorf("ID de transacción inválido: %d", transactionID)
	}

	return s.repo.VoidTransaction(ctx, transactionID)
}
