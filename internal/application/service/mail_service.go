package service

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// MailService handles sending emails to clients.
type MailServiceImpl struct {
	// Optional dialer factory for testing
	dialerFactory func(host string, port int, username, password string) Dialer
}

// Dialer abstracts gomail.Dialer for testing
type Dialer interface {
	DialAndSend(m ...*gomail.Message) error
}

// RealDialer wraps gomail.Dialer
type RealDialer struct {
	d *gomail.Dialer
}

func (r *RealDialer) DialAndSend(m ...*gomail.Message) error {
	return r.d.DialAndSend(m...)
}

// NewMailService creates a new instance of MailServiceImpl.
func NewMailService() *MailServiceImpl {
	return &MailServiceImpl{
		dialerFactory: func(host string, port int, username, password string) Dialer {
			return &RealDialer{d: gomail.NewDialer(host, port, username, password)}
		},
	}
}

// SendReceipt sends the authorized XML and RIDE PDF to the recipient.
func (s *MailServiceImpl) SendReceipt(issuer *domain.Issuer, recipientEmail string, xmlPath string, pdfPath string) error {
	if issuer.SMTPServer == nil || *issuer.SMTPServer == "" {
		return fmt.Errorf("SMTP configuration not found for issuer")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", *issuer.SMTPUser)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", fmt.Sprintf("Comprobante Electrónico - %s", issuer.BusinessName))
	
	body := fmt.Sprintf(`
	Estimado cliente,
	
	Adjunto encontrará su comprobante electrónico emitido por %s.
	
	Atentamente,
	%s
	`, issuer.BusinessName, issuer.TradeName)
	
	m.SetBody("text/plain", body)
	
	// Attachments
	m.Attach(xmlPath)
	m.Attach(pdfPath)

	// Dialer configuration
	port := 587
	if issuer.SMTPPort != nil {
		port = *issuer.SMTPPort
	}
	
	// Use factory to get dialer
	d := s.dialerFactory(*issuer.SMTPServer, port, *issuer.SMTPUser, *issuer.SMTPPassword)
	
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}
