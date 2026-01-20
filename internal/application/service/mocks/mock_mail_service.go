package mocks

import (
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) SendReceipt(issuer *domain.Issuer, recipientEmail string, receipt *domain.ElectronicReceipt, xmlPath string, pdfPath string) error {
	args := m.Called(issuer, recipientEmail, receipt, xmlPath, pdfPath)
	return args.Error(0)
}
