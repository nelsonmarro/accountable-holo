package sri

import "encoding/xml"

// InfoTributaria contiene la información común de todos los comprobantes
type InfoTributaria struct {
	XMLName         xml.Name `xml:"infoTributaria"`
	Ambiente        string   `xml:"ambiente"`
	TipoEmision     string   `xml:"tipoEmision"`
	RazonSocial     string   `xml:"razonSocial"`
	NombreComercial string   `xml:"nombreComercial,omitempty"`
	Ruc             string   `xml:"ruc"`
	ClaveAcceso     string   `xml:"claveAcceso"`
	CodDoc          string   `xml:"codDoc"`
	Estab           string   `xml:"estab"`
	PtoEmi          string   `xml:"ptoEmi"`
	Secuencial      string   `xml:"secuencial"`
	DirMatriz       string   `xml:"dirMatriz"`
}

// TotalImpuesto representa el desglose de un impuesto a nivel global
type TotalImpuesto struct {
	Codigo           string `xml:"codigo"`
	CodigoPorcentaje string `xml:"codigoPorcentaje"`
	BaseImponible    string `xml:"baseImponible"`
	Valor            string `xml:"valor"`
}

// Pago representa la forma de pago utilizada
type Pago struct {
	FormaPago    string `xml:"formaPago"`
	Total        string `xml:"total"`
	Plazo        string `xml:"plazo,omitempty"`
	UnidadTiempo string `xml:"unidadTiempo,omitempty"`
}

// InfoFactura contiene la información específica de una factura
type InfoFactura struct {
	FechaEmision                string `xml:"fechaEmision"`
	DirEstablecimiento          string `xml:"dirEstablecimiento,omitempty"`
	ContribuyenteEspecial       string `xml:"contribuyenteEspecial,omitempty"`
	ObligadoContabilidad        string `xml:"obligadoContabilidad,omitempty"`
	TipoIdentificacionComprador string `xml:"tipoIdentificacionComprador"`
	RazonSocialComprador        string `xml:"razonSocialComprador"`
	IdentificacionComprador     string `xml:"identificacionComprador"`
	DireccionComprador          string `xml:"direccionComprador,omitempty"`
	TotalSinImpuestos           string `xml:"totalSinImpuestos"`
	TotalDescuento              string `xml:"totalDescuento"`

	TotalConImpuestos struct {
		TotalImpuesto []TotalImpuesto `xml:"totalImpuesto"`
	} `xml:"totalConImpuestos"`

	Propina      string `xml:"propina"`
	ImporteTotal string `xml:"importeTotal"`
	Moneda       string `xml:"moneda"`

	Pagos struct {
		Pago []Pago `xml:"pago"`
	} `xml:"pagos"`
}

// Impuesto representa un impuesto aplicado a un detalle individual
type Impuesto struct {
	Codigo           string `xml:"codigo"`
	CodigoPorcentaje string `xml:"codigoPorcentaje"`
	Tarifa           string `xml:"tarifa"`
	BaseImponible    string `xml:"baseImponible"`
	Valor            string `xml:"valor"`
}

// Detalle representa una línea de producto o servicio en la factura
type Detalle struct {
	CodigoPrincipal        string `xml:"codigoPrincipal,omitempty"`
	Descripcion            string `xml:"descripcion"`
	Cantidad               string `xml:"cantidad"`
	PrecioUnitario         string `xml:"precioUnitario"`
	Descuento              string `xml:"descuento"`
	PrecioTotalSinImpuesto string `xml:"precioTotalSinImpuesto"`
	Impuestos              struct {
		Impuesto []Impuesto `xml:"impuesto"`
	} `xml:"impuestos"`
}

// Factura representa la estructura completa del XML de una factura v2.1.0
type Factura struct {
	XMLName xml.Name `xml:"factura"`
	ID      string   `xml:"id,attr"`      // Siempre "comprobante"
	Version string   `xml:"version,attr"` // "2.1.0"

	InfoTributaria InfoTributaria `xml:"infoTributaria"`
	InfoFactura    InfoFactura    `xml:"infoFactura"`
	Detalles       struct {
		Detalle []Detalle `xml:"detalle"`
	} `xml:"detalles"`
}