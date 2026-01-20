package mocks

import (
	"context"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockEmissionPointRepository struct {
	mock.Mock
}

func (m *MockEmissionPointRepository) GetByPoint(ctx context.Context, issuerID int, estCode, pointCode, receiptType string) (*domain.EmissionPoint, error) {
	args := m.Called(ctx, issuerID, estCode, pointCode, receiptType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EmissionPoint), args.Error(1)
}

func (m *MockEmissionPointRepository) IncrementSequence(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmissionPointRepository) Create(ctx context.Context, ep *domain.EmissionPoint) error {
	args := m.Called(ctx, ep)
	return args.Error(0)
}

func (m *MockEmissionPointRepository) GetAllByIssuer(ctx context.Context, issuerID int) ([]domain.EmissionPoint, error) {
	args := m.Called(ctx, issuerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.EmissionPoint), args.Error(1)
}

func (m *MockEmissionPointRepository) Update(ctx context.Context, ep *domain.EmissionPoint) error {
	args := m.Called(ctx, ep)
	return args.Error(0)
}
