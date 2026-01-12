package service

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"github.com/nelsonmarro/accountable-holo/config"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// MailService handles sending emails to clients.
type MailServiceImpl struct {
	// Optional dialer factory for testing
	dialerFactory func(host string, port int, username, password string) Dialer
	config        *config.Config
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
func NewMailService(cfg *config.Config) *MailServiceImpl {
	return &MailServiceImpl{
		config: cfg,
		dialerFactory: func(host string, port int, username, password string) Dialer {
			return &RealDialer{d: gomail.NewDialer(host, port, username, password)}
		},
	}
}

// SendReceipt sends the authorized XML and RIDE PDF to the recipient.
func (s *MailServiceImpl) SendReceipt(issuer *domain.Issuer, recipientEmail string, xmlPath string, pdfPath string) error {
	smtpCfg := s.config.SMTP
	if smtpCfg.Host == "" || smtpCfg.Port == 0 {
		return fmt.Errorf("SMTP configuration not found in app config")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpCfg.User) // Use configured user as sender
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

	// Use factory to get dialer
	d := s.dialerFactory(smtpCfg.Host, smtpCfg.Port, smtpCfg.User, smtpCfg.Password)
	
	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	return nil
}
