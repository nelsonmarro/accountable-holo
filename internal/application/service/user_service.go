package service

import (
	"context"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id int) error
	GetAllUsers(ctx context.Context) ([]domain.User, error)
}

// UserServiceImpl implements the UserService interface.
type UserServiceImpl struct {
	repo UserRepository
}

// NewUserService creates a new instance of UserServiceImpl.
func NewUserService(repo UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

// Login authenticates a user.
func (s *UserServiceImpl) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("credenciales incorrectas")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("credenciales incorrectas")
	}

	return user, nil
}

// CreateUser creates a new user.
func (s *UserServiceImpl) CreateUser(ctx context.Context, username, password, firstName, lastName string, role domain.UserRole, currentUser *domain.User) error {
	if currentUser.Role != domain.AdminRole {
		return fmt.Errorf("unauthorized: only admins can create users")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
	}

	return s.repo.CreateUser(ctx, user)
}

// UpdateUser updates an existing user.
func (s *UserServiceImpl) UpdateUser(ctx context.Context, id int, username, password, firstName, lastName string, role domain.UserRole, currentUser *domain.User) error {
	if currentUser.Role != domain.AdminRole {
		return fmt.Errorf("unauthorized: only admins can update users")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Username == "admin" {
		return fmt.Errorf("cannot update the default admin user")
	}

	user.Username = username
	user.FirstName = firstName
	user.LastName = lastName
	user.Role = role

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	return s.repo.UpdateUser(ctx, user)
}

// DeleteUser deletes a user.
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id int, currentUser *domain.User) error {
	if currentUser.Role != domain.AdminRole {
		return fmt.Errorf("unauthorized: only admins can delete users")
	}

	if currentUser.ID == id {
		return fmt.Errorf("cannot delete currently logged in user")
	}

	userToDelete, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if userToDelete.Username == "admin" {
		return fmt.Errorf("cannot delete the default admin user")
	}

	return s.repo.DeleteUser(ctx, id)
}

// GetAllUsers retrieves all users.
func (s *UserServiceImpl) GetAllUsers(ctx context.Context, currentUser *domain.User) ([]domain.User, error) {
	if currentUser.Role != domain.AdminRole {
		return nil, fmt.Errorf("unauthorized: only admins can view all users")
	}

	return s.repo.GetAllUsers(ctx)
}
