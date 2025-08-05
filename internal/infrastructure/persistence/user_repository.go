package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// UserRepositoryImpl implements the UserRepository interface.
type UserRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new instance of UserRepositoryImpl.
func NewUserRepository(db *pgxpool.Pool) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// CreateUser creates a new user in the database.
func (r *UserRepositoryImpl) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, password_hash, role, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	now := time.Now()
	err := r.db.QueryRow(ctx, query, user.Username, user.PasswordHash, user.Role, now, now).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByUsername retrieves a user from the database by their username.
func (r *UserRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetUserByID retrieves a user from the database by their ID.
func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// UpdateUser updates an existing user in the database.
func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET username = $1, password_hash = $2, role = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.Exec(ctx, query, user.Username, user.PasswordHash, user.Role, time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUser deletes a user from the database.
func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// GetAllUsers retrieves all users from the database.
func (r *UserRepositoryImpl) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, username, role, created_at, updated_at FROM users ORDER BY username ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
