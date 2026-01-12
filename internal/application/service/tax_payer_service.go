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

func (s *TaxPayerService) Search(ctx context.Context, query string) ([]domain.TaxPayer, error) {
	// Por ahora devolvemos todos, idealmente el repo debería tener Search(query)
	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	// Filtrado en memoria si el repo no soporta search (MVP)
	// O simplemente retornamos todos para el buscador de UI
	return all, nil
}
