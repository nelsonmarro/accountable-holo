// Package mocks provides mock implementations for testing purposes.
package mocks

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockAccountService is a mock type for the AccountService interface.
// It's defined here for testing purposes.
type MockAccountService struct {
	mock.Mock
}

// CreateNewAccount is the mock implementation of the interface method.
func (m *MockAccountService) CreateNewAccount(ctx context.Context, acc *domain.Account) error {
	// Record the method call and its arguments.
	args := m.Called(ctx, acc)
	// Return the pre-programmed error value (or nil).
	return args.Error(0)
}
