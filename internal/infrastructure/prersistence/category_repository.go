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

func (r *CategoryRepositoryImpl) GetPaginatedCategory(ctx context.Context, page, pageSize int) (
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

	var totalCount int64
	countQuery := `select count(*) from categories`
	err := r.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total category count: %w", err)
	}

	// there are no records, return empty result
	if totalCount == 0 {
		return &domain.PaginatedResult[domain.Category]{
			Data:       []domain.Category{},
			TotalCount: 0,
			Page:       1,
			PageSize:   0,
		}, nil
	}

	// calculate offset
	offset := (page - 1) * pageSize
	dataQuery := `
	   select id, name, type 
	   from categories 
		 order by name asc 
		 limit $1 offset $2
	`

	rows, err := r.db.Query(ctx, dataQuery, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated categories: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category rows: %w", err)
	}

	result := &domain.PaginatedResult[domain.Category]{
		Data:       categories,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	return result, nil
}

func (r *CategoryRepositoryImpl) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	query := `select id, name, type, created_at, updated_at 
	          from categories 
	          where id = $1`

	var cat domain.Category
	row := r.db.QueryRow(ctx, query, id)
	if err := row.Scan(&cat.ID, &cat.Name, &cat.Type, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}
	return &cat, nil
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

func (r *CategoryRepositoryImpl) UpdateCategory(ctx context.Context, category *domain.Category) error {
	query := `update categories 
	          set name = $1, type = $2, updated_at = $3 
	          where id = $4`
	now := time.Now()

	_, err := r.db.Exec(ctx, query, category.Name, category.Type, now, category.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	category.UpdatedAt = now

	return nil
}

func (r *CategoryRepositoryImpl) DeleteCategory(ctx context.Context, id int) error {
	query := `delete from categories where id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
