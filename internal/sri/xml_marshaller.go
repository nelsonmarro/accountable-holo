package sri

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// MarshalFactura convierte la estructura Factura en un arreglo de bytes XML listo para firmar.
// Usa xml.Marshal (compacto) para evitar problemas de integridad en la firma digital.
func MarshalFactura(factura *Factura) ([]byte, error) {
	if factura.ID != "comprobante" {
		factura.ID = "comprobante"
	}

	if factura.Version == "" {
		factura.Version = "1.1.0"
	}

	// Usamos Marshal (compacto) en lugar de MarshalIndent para el XML legal.
	// Esto reduce el riesgo de errores de "Firma Inválida" por espacios en blanco.
	xmlBytes, err := xml.Marshal(factura)
	if err != nil {
		return nil, fmt.Errorf("error al serializar la factura: %w", err)
	}

	// Construir el buffer con el encabezado estándar del SRI sin saltos de línea adicionales.
	var buffer bytes.Buffer
	buffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buffer.Write(xmlBytes)

	return buffer.Bytes(), nil
}