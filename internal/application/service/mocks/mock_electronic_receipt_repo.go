package mocks

import (
	"context"
	"time"

	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockElectronicReceiptRepository struct {
	mock.Mock
}

func (m *MockElectronicReceiptRepository) Create(ctx context.Context, er *domain.ElectronicReceipt) error {
	args := m.Called(ctx, er)
	return args.Error(0)
}

func (m *MockElectronicReceiptRepository) Update(ctx context.Context, er *domain.ElectronicReceipt) error {
	args := m.Called(ctx, er)
	return args.Error(0)
}

func (m *MockElectronicReceiptRepository) UpdateStatus(ctx context.Context, accessKey string, status string, message string, authDate *time.Time) error {
	args := m.Called(ctx, accessKey, status, message, authDate)
	return args.Error(0)
}

func (m *MockElectronicReceiptRepository) UpdateXML(ctx context.Context, accessKey string, xmlContent string) error {
	args := m.Called(ctx, accessKey, xmlContent)
	return args.Error(0)
}

func (m *MockElectronicReceiptRepository) UpdateTaxPayerID(ctx context.Context, accessKey string, taxPayerID int) error {
	args := m.Called(ctx, accessKey, taxPayerID)
	return args.Error(0)
}

func (m *MockElectronicReceiptRepository) UpdateEmailSent(ctx context.Context, accessKey string, sent bool) error {
	args := m.Called(ctx, accessKey, sent)
	return args.Error(0)
}

func (m *MockElectronicReceiptRepository) GetByAccessKey(ctx context.Context, accessKey string) (*domain.ElectronicReceipt, error) {
	args := m.Called(ctx, accessKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ElectronicReceipt), args.Error(1)
}

func (m *MockElectronicReceiptRepository) FindPendingReceipts(ctx context.Context) ([]domain.ElectronicReceipt, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ElectronicReceipt), args.Error(1)
}
