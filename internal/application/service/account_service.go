// Package service provides the implementations of business logic.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type AccountServiceImpl struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountServiceImpl {
	return &AccountServiceImpl{repo: repo}
}

func (s *AccountServiceImpl) CreateNewAccount(ctx context.Context, acc *domain.Account) error {
	validator := validator.New().For(acc)
	validator.Required("Name", "Type", "InitialBalance", "Number")
	validator.NumberMin(0, "InitialBalance")

	err := validator.ConsolidateErrors()
	if err != nil {
		return err
	}

	exists, err := s.repo.AccountExists(ctx, acc.Name, acc.Number, 0)
	if err != nil {
		return fmt.Errorf("error al verificar si la cuenta existe: %w", err)
	}

	if !exists {
		return errors.New("ya existe una cuenta con el mismo nombre o numero que la que trata de crear/nIntente otra vez")
	}

	err = s.repo.CreateAccount(ctx, acc)
	if err != nil {
		return fmt.Errorf("error al crear la cuenta: %w", err)
	}

	return nil
}

func (s *AccountServiceImpl) GetAllAccounts(ctx context.Context) ([]domain.Account, error) {
	return s.repo.GetAllAccounts(ctx)
}

func (s *AccountServiceImpl) GetAccountByID(ctx context.Context, id int) (*domain.Account, error) {
	if id <= 0 {
		return nil, errors.New("ID de cuenta inválido")
	}

	return s.repo.GetAccountByID(ctx, id)
}

func (s *AccountServiceImpl) UpdateAccount(ctx context.Context, acc *domain.Account) error {
	if acc.ID <= 0 {
		return errors.New("ID de cuenta inválido")
	}

	return s.repo.UpdateAccount(ctx, acc)
}

func (s *AccountServiceImpl) DeleteAccount(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("ID de cuenta inválido")
	}

	err := s.repo.DeleteAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("error al eliminar la cuenta: %w", err)
	}

	return nil
}
