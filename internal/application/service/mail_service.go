package service

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/nelsonmarro/verith/config"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/sri"
	"github.com/resend/resend-go/v3"
)

// MailServiceImpl handles sending emails using Resend.
type MailServiceImpl struct {
	client *resend.Client
	config *config.Config
}

// NewMailService creates a new instance of MailServiceImpl using Resend.
// Note: The 'mailPass' argument is kept for compatibility but ignored if using API Key from config.
func NewMailService(cfg *config.Config, mailPass string) *MailServiceImpl {
	apiKey := cfg.Email.APIKey
	// Si se pasa una clave "encriptada" (mailPass) y no hay en config, úsala (para transición suave)
	if apiKey == "" && mailPass != "" {
		apiKey = mailPass
	}

	client := resend.NewClient(apiKey)
	return &MailServiceImpl{
		client: client,
		config: cfg,
	}
}

// SendReceipt sends the authorized XML and RIDE PDF to the recipient using Resend API.
func (s *MailServiceImpl) SendReceipt(issuer *domain.Issuer, recipientEmail string, receipt *domain.ElectronicReceipt, xmlPath string, pdfPath string) error {
	// 1. Prepare Subject and Template Data
	subject := fmt.Sprintf("Comprobante Electrónico - %s", issuer.TradeName)
	clientName := "Cliente" // Could be fetched from receipt relations if loaded

	templatePath := "assets/templates/receipt_email.html"
	data := map[string]interface{}{
		"TradeName":  issuer.TradeName,
		"ClientName": clientName,
		"Year":       time.Now().Year(),
		"HasLogo":    false, // Resend doesn't support CID embedding easily, simplifying for now
	}

	if receipt.ReceiptType == "04" {
		templatePath = "assets/templates/credit_note_email.html"
		subject = fmt.Sprintf("Nota de Crédito Electrónica - %s", issuer.TradeName)

		// Extract NC specific data
		var nc sri.NotaCredito
		if err := xml.Unmarshal([]byte(receipt.XMLContent), &nc); err == nil {
			data["Motivo"] = nc.InfoNotaCredito.Motivo
			data["DocModificado"] = nc.InfoNotaCredito.NumDocModificado
		}
	}

	// 2. Render Template
	var body bytes.Buffer
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		// Fallback
		body.WriteString(fmt.Sprintf("Estimado cliente, adjunto su comprobante emitido por %s.", issuer.TradeName))
	} else {
		if err := t.Execute(&body, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
	}

	// 3. Prepare Attachments
	// Read file contents to memory
	xmlBytes, err := os.ReadFile(xmlPath)
	if err != nil {
		return fmt.Errorf("failed to read XML attachment: %w", err)
	}
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF attachment: %w", err)
	}

	attachments := []*resend.Attachment{
		{
			Filename: fmt.Sprintf("%s.xml", receipt.AccessKey),
			Content:  xmlBytes,
		},
		{
			Filename: fmt.Sprintf("%s.pdf", receipt.AccessKey),
			Content:  pdfBytes,
		},
	}

	// 4. Send Email via Resend
	// Use configured sender or fallback to testing sender
	from := s.config.Email.From
	if from == "" {
		from = "onboarding@resend.dev"
	}

	params := &resend.SendEmailRequest{
		From:        from,
		To:          []string{recipientEmail},
		Subject:     subject,
		Html:        body.String(),
		Attachments: attachments,
	}

	_, err = s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend api error: %w", err)
	}

	return nil
}
