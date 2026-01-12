package sri

import (
	"fmt"
	"os"

	"github.com/nelsonmarro/go_ec_sri_invoice_signer/pkg/signer"
)

// DocumentSigner es el servicio encargado de firmar digitalmente los comprobantes.
type DocumentSigner struct {
	p12Path  string
	password string
}

// NewDocumentSigner crea una nueva instancia del firmador con la ruta del certificado y su contraseña.
func NewDocumentSigner(p12Path string, password string) *DocumentSigner {
	return &DocumentSigner{
		p12Path:  p12Path,
		password: password,
	}
}

// Sign toma el XML crudo y devuelve el XML firmado usando el paquete especializado.
func (s *DocumentSigner) Sign(xmlBytes []byte) ([]byte, error) {
	// 1. Leer el archivo de firma (.p12)
	p12Bytes, err := os.ReadFile(s.p12Path)
	if err != nil {
		return nil, fmt.Errorf("error al leer el certificado .p12: %w", err)
	}

	// 2. Firmar usando la librería propia del usuario
	signedXML, err := signer.SignInvoice(string(xmlBytes), p12Bytes, &signer.SignOptions{
		Password: s.password,
	})
	if err != nil {
		return nil, fmt.Errorf("error al firmar el XML: %w", err)
	}

	return []byte(signedXML), nil
}
