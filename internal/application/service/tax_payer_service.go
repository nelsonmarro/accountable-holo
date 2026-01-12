package service

import (
	"context"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TaxPayerService struct {
	repo TaxPayerRepository
}

func NewTaxPayerService(repo TaxPayerRepository) *TaxPayerService {
	return &TaxPayerService{repo: repo}
}

func (s *TaxPayerService) GetByID(ctx context.Context, id int) (*domain.TaxPayer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaxPayerService) GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error) {
	return s.repo.GetByIdentification(ctx, identification)
}

func (s *TaxPayerService) Create(ctx context.Context, tp *domain.TaxPayer) error {
	// Validaciones básicas
	if len(tp.Identification) < 10 {
		return fmt.Errorf("la identificación debe tener al menos 10 dígitos")
	}
	if tp.Name == "" || tp.Email == "" {
		return fmt.Errorf("nombre y email son obligatorios")
	}
	return s.repo.Create(ctx, tp)
}

func (s *TaxPayerService) Update(ctx context.Context, tp *domain.TaxPayer) error {
	if tp.ID == 0 {
		return fmt.Errorf("ID inválido para actualización")
	}
	return s.repo.Update(ctx, tp)
}

func (s *TaxPayerService) GetPaginated(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResult[domain.TaxPayer], error) {
	return s.repo.GetPaginated(ctx, page, pageSize, search)
}

func (s *TaxPayerService) Search(ctx context.Context, query string) ([]domain.TaxPayer, error) {
	// Deprecated: Use GetPaginated instead for UI
	return s.repo.GetAll(ctx)
}
