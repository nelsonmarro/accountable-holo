package mocks

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, username, password, firstName, lastName string, role domain.UserRole, currentUser *domain.User) error {
	args := m.Called(ctx, username, password, firstName, lastName, role, currentUser)
	return args.Error(0)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int, username, password, firstName, lastName string, role domain.UserRole, currentUser *domain.User) error {
	args := m.Called(ctx, id, username, password, firstName, lastName, role, currentUser)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id int, currentUser *domain.User) error {
	args := m.Called(ctx, id, currentUser)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers(ctx context.Context, currentUser *domain.User) ([]domain.User, error) {
	args := m.Called(ctx, currentUser)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserService) GetAdminUsers(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserService) ResetPassword(ctx context.Context, username, newPassword string) error {
	args := m.Called(ctx, username, newPassword)
	return args.Error(0)
}

func (m *MockUserService) HasUsers(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}
