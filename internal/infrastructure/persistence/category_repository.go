package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type CategoryRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{db: db}
}

// getPaginatedCategories is the private helper function that contains the common logic.
func (r *CategoryRepositoryImpl) getPaginatedCategories(
	ctx context.Context,
	page, pageSize int,
	baseWhereClause string,
	filter ...string,
) (*domain.PaginatedResult[domain.Category], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if page > 100 {
		page = 100
	}

	var queryArgs []any
	var countQueryArgs []any
	whereClauses := []string{}

	// Start with the base WHERE clause if it exists
	if baseWhereClause != "" {
		whereClauses = append(whereClauses, baseWhereClause)
	}

	// Add the user's search filter if provided
	if len(filter) > 0 && filter[0] != "" {
		// The parameter index will depend on how many clauses we already have
		paramIndex := len(queryArgs) + 1
		filterClause := fmt.Sprintf("(name ILIKE $%d OR type ILIKE $%d)", paramIndex, paramIndex)
		whereClauses = append(whereClauses, filterClause)

		filterArg := "%" + filter[0] + "%"
		queryArgs = append(queryArgs, filterArg)
		countQueryArgs = append(countQueryArgs, filterArg)
	}

	// Join all WHERE clauses with " AND "
	fullWhereClause := ""
	if len(whereClauses) > 0 {
		fullWhereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// --- Count Query ---
	countQuery := `SELECT count(*) FROM categories` + fullWhereClause
	var totalCount int64
	err := r.db.QueryRow(ctx, countQuery, countQueryArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total category count: %w", err)
	}

	if totalCount == 0 {
		return &domain.PaginatedResult[domain.Category]{
			Data:       []domain.Category{},
			TotalCount: 0,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}

	// --- Data Query ---
	offset := (page - 1) * pageSize
	limitParamIndex := len(queryArgs) + 1
	offsetParamIndex := len(queryArgs) + 2

	limitOffsetClause := fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", limitParamIndex, offsetParamIndex)
	finalQueryArgs := append(queryArgs, pageSize, offset)

	dataQuery := `SELECT id, name, type FROM categories` + fullWhereClause + limitOffsetClause

	rows, err := r.db.Query(ctx, dataQuery, finalQueryArgs...)
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

	return &domain.PaginatedResult[domain.Category]{
		Data:       categories,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetPaginatedCategories calls the helper without any base WHERE clause.
func (r *CategoryRepositoryImpl) GetPaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
	*domain.PaginatedResult[domain.Category],
	error,
) {
	return r.getPaginatedCategories(ctx, page, pageSize, "", filter...)
}

// GetSelectablePaginatedCategories calls the helper with the clause to exclude system categories.
func (r *CategoryRepositoryImpl) GetSelectablePaginatedCategories(ctx context.Context, page, pageSize int, filter ...string) (
	*domain.PaginatedResult[domain.Category],
	error,
) {
	baseWhere := "name NOT LIKE '%Anular Transacción%'"
	return r.getPaginatedCategories(ctx, page, pageSize, baseWhere, filter...)
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

func (r *CategoryRepositoryImpl) CategoryExists(ctx context.Context, name string, id int) (bool, error) {
	query := `select exists(select 1 from categories where name = $1 and id != $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, name, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if category exists: %w", err)
	}
	return exists, nil
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

func (r *CategoryRepositoryImpl) FindByNameAndType(ctx context.Context, name string, catType domain.CategoryType) (*domain.Category, error) {
	query := `SELECT id, name, type, created_at, updated_at FROM categories WHERE name = $1 AND type = $2`
	var category domain.Category
	err := r.db.QueryRow(ctx, query, name, catType).Scan(&category.ID, &category.Name, &category.Type, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category with name '%s' and type '%s' not found", name, catType)
		}
		return nil, fmt.Errorf("failed to scan category by name and type: %w", err)
	}
	return &category, nil
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
