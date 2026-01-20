package sri

import (
	"fmt"
	"os"

	"github.com/nelsonmarro/go_ec_sri_invoice_signer/pkg/signer"
)

// Re-exportamos los tipos de la librería para facilitar su uso
const (
	SHA1   = signer.SHA1
	SHA256 = signer.SHA256
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
func (s *DocumentSigner) Sign(xmlBytes []byte, algo signer.HashAlgorithm) ([]byte, error) {
	// 1. Leer el archivo de firma (.p12)
	p12Bytes, err := os.ReadFile(s.p12Path)
	if err != nil {
		return nil, fmt.Errorf("error al leer el certificado .p12: %w", err)
	}

	// 2. Firmar usando la librería propia
	signedXML, err := signer.SignInvoice(string(xmlBytes), p12Bytes, &signer.SignOptions{
		Password:  s.password,
		Algorithm: algo,
	})
	if err != nil {
		return nil, fmt.Errorf("error al firmar el XML: %w", err)
	}

	return []byte(signedXML), nil
}

// SignCreditNote firma un XML de Nota de Crédito.
func (s *DocumentSigner) SignCreditNote(xmlBytes []byte, algo signer.HashAlgorithm) ([]byte, error) {
	// 1. Leer el archivo de firma (.p12)
	p12Bytes, err := os.ReadFile(s.p12Path)
	if err != nil {
		return nil, fmt.Errorf("error al leer el certificado .p12: %w", err)
	}

	// 2. Firmar usando la librería propia (Método específico para NC)
	signedXML, err := signer.SignCreditNote(string(xmlBytes), p12Bytes, &signer.SignOptions{
		Password:  s.password,
		Algorithm: algo,
	})
	if err != nil {
		return nil, fmt.Errorf("error al firmar la Nota de Crédito: %w", err)
	}

	return []byte(signedXML), nil
}