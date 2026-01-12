package sri

import (
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type RideGenerator struct{}

func NewRideGenerator() *RideGenerator {
	return &RideGenerator{}
}

func (g *RideGenerator) GenerateFacturaRide(factura *Factura, ridePath string, logoPath string, authDate time.Time, authNumber string) error {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithOrientation(orientation.Vertical).
		Build()

	m := maroto.New(cfg)

	// --- 1. Header ---
	m.AddRows(g.buildHeader(factura, logoPath, authDate, authNumber))

	// --- 2. Client Info ---
	m.AddRows(g.buildClientInfo(factura))

	// --- 3. Details ---
	m.AddRows(g.buildDetailsRows(factura)...)

	// --- 4. Footer ---
	m.AddRows(g.buildFooter(factura))

	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("error generando RIDE: %w", err)
	}

	return document.Save(ridePath)
}

func (g *RideGenerator) buildHeader(f *Factura, logoPath string, authDate time.Time, authNumber string) core.Row {
	boldStyle := props.Text{Style: fontstyle.Bold, Size: 10, Align: align.Left}
	normalStyle := props.Text{Size: 9, Align: align.Left}
	titleStyle := props.Text{Style: fontstyle.Bold, Size: 14, Align: align.Left}

	ambiente := "PRUEBAS"
	if f.InfoTributaria.Ambiente == "2" {
		ambiente = "PRODUCCIÓN"
	}

	leftCol := col.New(6).Add(
		text.New(f.InfoTributaria.RazonSocial, titleStyle),
		text.New(f.InfoTributaria.NombreComercial, boldStyle),
		text.New("Dirección Matriz:", boldStyle),
		text.New(f.InfoTributaria.DirMatriz, normalStyle),
		text.New("Dirección Sucursal:", boldStyle),
		text.New(f.InfoFactura.DirEstablecimiento, normalStyle),
		text.New("Obligado a Llevar Contabilidad: "+f.InfoFactura.ObligadoContabilidad, normalStyle),
	)

	rightCol := col.New(6).Add(
		text.New("R.U.C.: "+f.InfoTributaria.Ruc, titleStyle),
		text.New("FACTURA", titleStyle),
		text.New(fmt.Sprintf("No. %s-%s-%s", f.InfoTributaria.Estab, f.InfoTributaria.PtoEmi, f.InfoTributaria.Secuencial), boldStyle),
		text.New("NÚMERO DE AUTORIZACIÓN", boldStyle),
		text.New(authNumber, normalStyle),
		text.New("FECHA Y HORA DE AUTORIZACIÓN", boldStyle),
		text.New(authDate.Format("2006-01-02 15:04:05"), normalStyle),
		text.New("AMBIENTE: "+ambiente, normalStyle),
		text.New("CLAVE DE ACCESO", boldStyle),
		code.NewBar(f.InfoTributaria.ClaveAcceso, props.Barcode{Center: false, Proportion: props.Proportion{Width: 20, Height: 2}}),
		text.New(f.InfoTributaria.ClaveAcceso, props.Text{Size: 8, Align: align.Center}),
		code.NewQr(f.InfoTributaria.ClaveAcceso, props.Rect{Center: true, Percent: 50}),
	)

	return row.New(115).Add(leftCol, rightCol)
}

func (g *RideGenerator) buildClientInfo(f *Factura) core.Row {
	styleVal := props.Text{Size: 9}
	return row.New(20).Add(
		col.New(12).Add(
			text.New("Razón Social / Nombres y Apellidos: "+f.InfoFactura.RazonSocialComprador, styleVal),
			text.New("Identificación: "+f.InfoFactura.IdentificacionComprador, styleVal),
			text.New("Fecha Emisión: "+f.InfoFactura.FechaEmision, styleVal),
		),
	)
}

func (g *RideGenerator) buildDetailsRows(f *Factura) []core.Row {
	var rows []core.Row
	headerRow := row.New(8).WithStyle(&props.Cell{BackgroundColor: &props.Color{Red: 230, Green: 230, Blue: 230}})
	headerRow.Add(
		col.New(2).Add(text.New("Cod.", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(1).Add(text.New("Cant.", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(5).Add(text.New("Descripción", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(2).Add(text.New("P. Unit.", props.Text{Style: fontstyle.Bold, Size: 8})),
		col.New(2).Add(text.New("Total", props.Text{Style: fontstyle.Bold, Size: 8})),
	)
	rows = append(rows, headerRow)

	for _, det := range f.Detalles.Detalle {
		r := row.New(6).Add(
			col.New(2).Add(text.New(det.CodigoPrincipal, props.Text{Size: 8})),
			col.New(1).Add(text.New(det.Cantidad, props.Text{Size: 8})),
			col.New(5).Add(text.New(det.Descripcion, props.Text{Size: 8})),
			col.New(2).Add(text.New(det.PrecioUnitario, props.Text{Size: 8, Align: align.Right})),
			col.New(2).Add(text.New(det.PrecioTotalSinImpuesto, props.Text{Size: 8, Align: align.Right})),
		)
		rows = append(rows, r)
	}
	return rows
}

func (g *RideGenerator) buildFooter(f *Factura) core.Row {
	leftCol := col.New(7).Add(
		text.New("Información Adicional", props.Text{Style: fontstyle.Bold, Size: 9, Top: 2}),
		text.New("Email: "+f.InfoFactura.DireccionComprador, props.Text{Size: 8}),
	)

	rightCol := col.New(5)
	// Simplified totals for compilation check
	rightCol.Add(
		text.New("SUBTOTAL 15%: "+f.InfoFactura.TotalSinImpuestos, props.Text{Size: 8}),
		text.New("IVA 15%: "+f.InfoFactura.ImporteTotal, props.Text{Size: 8}),
		text.New("VALOR TOTAL: "+f.InfoFactura.ImporteTotal, props.Text{Style: fontstyle.Bold, Size: 10}),
	)

	return row.New(40).Add(leftCol, rightCol)
}
