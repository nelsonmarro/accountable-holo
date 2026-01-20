package sri

import "encoding/xml"

// NotaCredito representa la estructura XML completa de una Nota de Crédito (Tipo 04)
type NotaCredito struct {
	XMLName         xml.Name        `xml:"notaCredito"`
	ID              string          `xml:"id,attr"`
	Version         string          `xml:"version,attr"`
	InfoTributaria  InfoTributaria  `xml:"infoTributaria"`
	InfoNotaCredito InfoNotaCredito `xml:"infoNotaCredito"`
	Detalles        DetallesNC      `xml:"detalles"`
}

type InfoNotaCredito struct {
	FechaEmision                string            `xml:"fechaEmision"`
	DirEstablecimiento          string            `xml:"dirEstablecimiento"`
	TipoIdentificacionComprador string            `xml:"tipoIdentificacionComprador"`
	RazonSocialComprador        string            `xml:"razonSocialComprador"`
	IdentificacionComprador     string            `xml:"identificacionComprador"`
	ContribuyenteEspecial       string            `xml:"contribuyenteEspecial,omitempty"`
	ObligadoContabilidad        string            `xml:"obligadoContabilidad"`
	Rise                        string            `xml:"rise,omitempty"`
	CodDocModificado            string            `xml:"codDocModificado"`            // Siempre "01" para facturas
	NumDocModificado            string            `xml:"numDocModificado"`            // 001-001-000000123
	FechaEmisionDocSustento     string            `xml:"fechaEmisionDocSustento"`     // Fecha de la factura original
	TotalSinImpuestos           string            `xml:"totalSinImpuestos"`
	ValorModificacion           string            `xml:"valorModificacion"`           // Monto total de la nota de crédito
	Moneda                      string            `xml:"moneda"`
	TotalConImpuestos           TotalConImpuestos `xml:"totalConImpuestos"`
	Motivo                      string            `xml:"motivo"`
}

// DetallesNC estructura contenedora para la lista de items de Nota de Crédito
type DetallesNC struct {
	Detalle []DetalleNC `xml:"detalle"`
}

// DetalleNC representa una línea de la Nota de Crédito.
// Difiere de la Factura en que usa 'codigoInterno' en lugar de 'codigoPrincipal'.
type DetalleNC struct {
	CodigoInterno          string `xml:"codigoInterno,omitempty"` // Diferente a Factura
	CodigoAdicional        string `xml:"codigoAdicional,omitempty"`
	Descripcion            string `xml:"descripcion"`
	Cantidad               string `xml:"cantidad"`
	PrecioUnitario         string `xml:"precioUnitario"`
	Descuento              string `xml:"descuento"`
	PrecioTotalSinImpuesto string `xml:"precioTotalSinImpuesto"`
	Impuestos              struct {
		Impuesto []Impuesto `xml:"impuesto"`
	} `xml:"impuestos"`
}
