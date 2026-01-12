package service

import (
	"errors"
	"testing"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/gomail.v2"
)

// MockDialer simulates email sending
type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(msg ...*gomail.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestSendReceipt(t *testing.T) {
	// Test Data
	server := "smtp.test.com"
	port := 587
	user := "sender@test.com"
	pass := "secret"
	
	issuer := &domain.Issuer{
		SMTPServer:   &server,
		SMTPPort:     &port,
		SMTPUser:     &user,
		SMTPPassword: &pass,
		BusinessName: "Test Corp",
		TradeName:    "Test Store",
	}

	t.Run("Success", func(t *testing.T) {
		mockDialer := new(MockDialer)
		mockDialer.On("DialAndSend", mock.Anything).Return(nil).Once()

		svc := &MailServiceImpl{
			dialerFactory: func(h string, p int, u, pwd string) Dialer {
				assert.Equal(t, server, h)
				assert.Equal(t, port, p)
				assert.Equal(t, user, u)
				assert.Equal(t, pass, pwd)
				return mockDialer
			},
		}

		err := svc.SendReceipt(issuer, "client@test.com", "factura.xml", "factura.pdf")
		assert.NoError(t, err)
		mockDialer.AssertExpectations(t)
	})

	t.Run("Fail - SMTP Config Missing", func(t *testing.T) {
		svc := NewMailService()
		emptyIssuer := &domain.Issuer{}
		
		err := svc.SendReceipt(emptyIssuer, "client@test.com", "x", "p")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP configuration not found")
	})

	t.Run("Fail - Send Error", func(t *testing.T) {
		mockDialer := new(MockDialer)
		mockDialer.On("DialAndSend", mock.Anything).Return(errors.New("connection timeout")).Once()

		svc := &MailServiceImpl{
			dialerFactory: func(h string, p int, u, pwd string) Dialer {
				return mockDialer
			},
		}

		err := svc.SendReceipt(issuer, "client@test.com", "x", "p")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection timeout")
	})
}
