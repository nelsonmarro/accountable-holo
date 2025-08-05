package user

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	CreateUser(ctx context.Context, username, password string, role domain.UserRole, currentUser *domain.User) error
	UpdateUser(ctx context.Context, id int, username, password string, role domain.UserRole, currentUser *domain.User) error
	DeleteUser(ctx context.Context, id int, currentUser *domain.User) error
	GetAllUsers(ctx context.Context, currentUser *domain.User) ([]domain.User, error)
}
