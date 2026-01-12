package mocks

import (
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockMailService struct {
	mock.Mock
}

func (m *MockMailService) SendReceipt(issuer *domain.Issuer, recipientEmail string, xmlPath string, pdfPath string) error {
	args := m.Called(issuer, recipientEmail, xmlPath, pdfPath)
	return args.Error(0)
}
