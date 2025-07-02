// Package persistence provides the persistence layer for the application.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type AccountRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepositoryImpl {
	return &AccountRepositoryImpl{db: db}
}

func (r *AccountRepositoryImpl) GetAllAccounts(ctx context.Context) ([]domain.Account, error) {
	query := `select id, name, number, type, initial_balance, created_at, updated_at 
	          from accounts
	          order by name asc`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.Account

	for rows.Next() {
		acc := domain.Account{}
		err := rows.Scan(&acc.ID, &acc.Name, &acc.Number, &acc.Type, &acc.InitialBalance, &acc.CreatedAt, &acc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account row: %w", err)
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over account rows: %w", err)
	}

	return accounts, nil
}

func (r *AccountRepositoryImpl) GetAccountByID(ctx context.Context, id int) (*domain.Account, error) {
	query := `select id, name, number, type, initial_balance, created_at, updated_at
	          from accounts
	          where id = $1`

	var acc domain.Account

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&acc.ID,
		&acc.Name,
		&acc.Number,
		&acc.Type,
		&acc.InitialBalance,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by id: %w", err)
	}

	return &acc, nil
}

func (r *AccountRepositoryImpl) AccountExists(ctx context.Context, name, number string, id int) (bool, error) {
	query := `select exists(select 1 from accounts where (name = $1 or number = $2) and id != $3)`
	var exists bool

	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if account exists: %w", err)
	}
	return exists, nil
}

func (r *AccountRepositoryImpl) CreateAccount(ctx context.Context, acc *domain.Account) error {
	query := `insert into accounts (name, number, type, initial_balance, created_at, updated_at) values ($1, $2, $3, $4, $5, $6) returning id, created_at, updated_at`

	err := r.db.QueryRow(
		ctx, query,
		acc.Name,
		acc.Number,
		acc.Type,
		acc.InitialBalance,
		time.Now(),
		time.Now(),
	).Scan(&acc.ID, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (r *AccountRepositoryImpl) DeleteAccount(ctx context.Context, id int) error {
	query := `delete from accounts where id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no account found with id %d", id)
	}

	return nil
}

func (r *AccountRepositoryImpl) UpdateAccount(ctx context.Context, acc *domain.Account) error {
	query := `update accounts set name = $1, number = $2, type = $3, updated_at = $4 where id = $5`

	result, err := r.db.Exec(
		ctx,
		query,
		acc.Name,
		acc.Number,
		acc.Type,
		time.Now(),
		acc.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no account found with id %d", acc.ID)
	}

	return nil
}
