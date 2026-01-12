package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Save(ctx context.Context, sourcePath string, destinationName string) (string, error) {
	args := m.Called(ctx, sourcePath, destinationName)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) GetFullPath(storagePath string) (string, error) {
	args := m.Called(storagePath)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) Delete(ctx context.Context, storagePath string) error {
	args := m.Called(ctx, storagePath)
	return args.Error(0)
}
