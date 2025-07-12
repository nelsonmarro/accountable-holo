package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type CategoryServiceImpl struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryServiceImpl {
	return &CategoryServiceImpl{repo: repo}
}

func (s *CategoryServiceImpl) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetAllCategories(ctx)
}

func (s *CategoryServiceImpl) GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
	*domain.PaginatedResult[domain.Category],
	error,
) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if page > 100 {
		page = 100
	}

	return s.repo.GetPaginatedCategories(ctx, page, pageSize, filter...)
}

func (s *CategoryServiceImpl) GetSelectablePaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
	*domain.PaginatedResult[domain.Category],
	error,
) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if page > 100 {
		page = 100
	}

	return s.repo.GetSelectablePaginatedCategories(ctx, page, pageSize, filter...)
}

func (s *CategoryServiceImpl) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	if id < 0 {
		return nil, fmt.Errorf("ID de cuenta inválido: %d ", id)
	}
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *CategoryServiceImpl) CreateCategory(ctx context.Context, category *domain.Category) error {
	catValidator := validator.New().For(category)
	catValidator.Required("Name", "Type")

	err := catValidator.ConsolidateErrors()
	if err != nil {
		return err
	}

	exists, err := s.repo.CategoryExists(ctx, category.Name, 0)
	if err != nil {
		return fmt.Errorf("error al verificar si la categoria existe: %w", err)
	}
	if exists {
		return errors.New("ya existe una categoria con el mismo nombre ingresado")
	}

	err = s.repo.CreateCategory(ctx, category)
	if err != nil {
		return err
	}

	return nil
}

func (s *CategoryServiceImpl) UpdateCategory(ctx context.Context, category *domain.Category) error {
	if category.ID < 0 {
		return fmt.Errorf("ID de cuenta inválido: %d ", category.ID)
	}

	catValidator := validator.New().For(category)
	catValidator.Required("Name", "Type")
	err := catValidator.ConsolidateErrors()
	if err != nil {
		return err
	}

	exists, err := s.repo.CategoryExists(ctx, category.Name, category.ID)
	if err != nil {
		return fmt.Errorf("error al verificar si la categoria existe: %w", err)
	}
	if exists {
		return errors.New("ya existe otra categoria con el mismo nombre ingresado")
	}

	err = s.repo.UpdateCategory(ctx, category)
	if err != nil {
		return err
	}
	return nil
}

func (s *CategoryServiceImpl) DeleteCategory(ctx context.Context, id int) error {
	if id < 0 {
		return fmt.Errorf("ID de cuenta inválido: %d ", id)
	}

	err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
