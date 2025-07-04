package service

import (
	"context"
	"fmt"

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
