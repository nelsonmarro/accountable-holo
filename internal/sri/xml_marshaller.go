package sri

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// MarshalFactura convierte la estructura Factura en un arreglo de bytes XML listo para firmar.                                       │
// Agrega automáticamente el encabezado XML requerido por el SRI.
func MarshalFactura(factura *Factura) ([]byte, error) {
	if factura.ID != "comprobante" {
		factura.ID = "comprobante"
	}

	if factura.Version != "2.1.0" {
		factura.Version = "2.1.0"
	}

	xmlBytes, err := xml.MarshalIndent(factura, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error al serializar la factura: %w", err)
	}

	// 3. Construir el buffer final con el encabezado explícito                                                                       │
	// El SRI exige UTF-8 y el encabezado estándar.
	var buffer bytes.Buffer
	buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buffer.Write(xmlBytes)

	return buffer.Bytes(), nil
}
