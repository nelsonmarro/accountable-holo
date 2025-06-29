package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type CategoryRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{db: db}
}

func (r *CategoryRepositoryImpl) CreateCategory(ctx context.Context, category *domain.Category) error {
	query := `insert into categories (name, type, created_at, updated_at) 
	                          values ($1, $2, $3, $4) 
	                          returning id, created_at, updated_at`

	now := time.Now()
	err := r.db.QueryRow(ctx, query, category.Name, category.Type, now, now).
		Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *CategoryRepositoryImpl) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	query := `select id, name, type 
	          from categories 
	          order by name asc`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Type); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}
