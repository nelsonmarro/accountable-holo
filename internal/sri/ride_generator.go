package sri

import (
	"fmt"
	"strconv"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
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
		WithLeftMargin(10).
		WithRightMargin(10).
		WithTopMargin(10).
		Build()

	m := maroto.New(cfg)

	// --- 1. Cabecera ---
	m.AddRows(g.buildHeader(factura, logoPath, authDate, authNumber)...)

	m.AddRow(2)

	// --- 2. Información del Cliente ---
	m.AddRows(g.buildClientInfo(factura)...)

	m.AddRow(2)

	// --- 3. Detalles de Productos ---
	m.AddRows(g.buildDetailsRows(factura)...)

	m.AddRow(2)

	// --- 4. Pie de Página ---
	// Usaremos una fila contenedora para dividir Izquierda (Info) y Derecha (Totales)
	// Pero como queremos filas reales para la tabla de totales, usaremos un truco:
	// Dos columnas, pero la columna derecha contendrá filas internas si Maroto v2 lo soporta (no lo hace bien).
	// Alternativa: Usar posicionamiento absoluto mejorado para la tabla derecha.
	m.AddRows(g.buildFooter(factura)...)

	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("error generando RIDE: %w", err)
	}

	return document.Save(ridePath)
}

func (g *RideGenerator) buildHeader(f *Factura, logoPath string, authDate time.Time, authNumber string) []core.Row {
	ambiente := "PRUEBAS"
	if f.InfoTributaria.Ambiente == "2" {
		ambiente = "PRODUCCIÓN"
	}

	var logo core.Component
	if logoPath != "" {
		logo = image.NewFromFile(logoPath)
	} else {
		logo = text.New("NO TIENE LOGO", props.Text{
			Style: fontstyle.Bold, Size: 16, Align: align.Center, Top: 10, Color: &props.Color{Red: 200},
		})
	}

	return []core.Row{
		row.New(90).Add(
			col.New(6).Add(
				logo,
				text.New(" ", props.Text{Top: 35}), // Spacer
				// Info Emisor
				text.New(f.InfoTributaria.RazonSocial, props.Text{Style: fontstyle.Bold, Size: 8, Top: 40}),
				text.New(f.InfoTributaria.NombreComercial, props.Text{Size: 8, Top: 45}),
				text.New("Dirección Matriz:", props.Text{Style: fontstyle.Bold, Size: 8, Top: 55}),
				text.New(f.InfoTributaria.DirMatriz, props.Text{Size: 8, Top: 60}),
				text.New("Dirección Sucursal:", props.Text{Style: fontstyle.Bold, Size: 8, Top: 68}),
				text.New(f.InfoFactura.DirEstablecimiento, props.Text{Size: 8, Top: 73}),
				text.New("OBLIGADO A LLEVAR CONTABILIDAD: "+f.InfoFactura.ObligadoContabilidad, props.Text{Size: 8, Top: 82}),
			),
			// Columna Derecha con Borde
			col.New(6).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
				text.New("R.U.C.: "+f.InfoTributaria.Ruc, props.Text{Style: fontstyle.Bold, Size: 10, Top: 3, Left: 2}),
				text.New("FACTURA", props.Text{Style: fontstyle.Bold, Size: 12, Top: 9, Left: 2}),
				text.New("No. "+fmt.Sprintf("%s-%s-%s", f.InfoTributaria.Estab, f.InfoTributaria.PtoEmi, f.InfoTributaria.Secuencial), props.Text{Size: 10, Top: 15, Left: 2}),
				text.New("NÚMERO DE AUTORIZACIÓN", props.Text{Style: fontstyle.Bold, Size: 8, Top: 22, Left: 2}),
				text.New(authNumber, props.Text{Size: 8, Top: 26, Left: 2}),
				text.New("FECHA Y HORA DE AUTORIZACIÓN", props.Text{Style: fontstyle.Bold, Size: 8, Top: 32, Left: 2}),
				text.New(authDate.Format("02/01/2006 15:04:05"), props.Text{Size: 8, Top: 36, Left: 2}),
				text.New("AMBIENTE: "+ambiente, props.Text{Size: 8, Top: 42, Left: 2}),
				text.New("EMISIÓN: NORMAL", props.Text{Size: 8, Top: 48, Left: 2}),
				text.New("CLAVE DE ACCESO", props.Text{Style: fontstyle.Bold, Size: 8, Top: 60, Left: 2}),

				// Código de barras posicionado al fondo absoluto
				code.NewBar(f.InfoTributaria.ClaveAcceso, props.Barcode{Percent: 67, Top: 68, Left: 16}),
				text.New(f.InfoTributaria.ClaveAcceso, props.Text{Size: 7, Align: align.Center, Top: 82}),
			),
		),
	}
}

func (g *RideGenerator) buildClientInfo(f *Factura) []core.Row {
	return []core.Row{
		row.New(22).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
			col.New(7).Add(
				text.New("Razón Social / Nombres y Apellidos: "+f.InfoFactura.RazonSocialComprador, props.Text{Style: fontstyle.Bold, Size: 7, Top: 2, Left: 2}),
				text.New("Identificación: "+f.InfoFactura.IdentificacionComprador, props.Text{Size: 7, Top: 8, Left: 2}),
				text.New("Fecha Emisión: "+f.InfoFactura.FechaEmision, props.Text{Size: 7, Top: 14, Left: 2}),
			),
			col.New(5).Add(
				text.New("Dirección: "+f.InfoFactura.DireccionComprador, props.Text{Size: 7, Top: 14}),
				text.New("Guía Remisión: ", props.Text{Size: 7, Top: 14, Align: align.Right, Right: 30}),
			),
		),
	}
}

func (g *RideGenerator) buildDetailsRows(f *Factura) []core.Row {
	var rows []core.Row
	headerStyle := &props.Cell{BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240}, BorderType: border.Full, BorderThickness: 0.1}

	rows = append(rows, row.New(8).WithStyle(headerStyle).Add(
		col.New(2).Add(text.New("Cod. Principal", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Center, Top: 1.5})),
		col.New(1).Add(text.New("Cant.", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Center, Top: 1.5})),
		col.New(5).Add(text.New("Descripción", props.Text{Style: fontstyle.Bold, Size: 7, Top: 1.5, Left: 2})),
		col.New(2).Add(text.New("Precio Unitario", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Right, Top: 1.5, Right: 2})),
		col.New(1).Add(text.New("Desc.", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Right, Top: 1.5, Right: 2})),
		col.New(1).Add(text.New("Precio Total", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Right, Top: 1.5, Right: 2})),
	))

	for _, det := range f.Detalles.Detalle {
		// Formatear cantidad y precio a 2 decimales para visualización limpia
		cant, _ := strconv.ParseFloat(det.Cantidad, 64)
		pUnit, _ := strconv.ParseFloat(det.PrecioUnitario, 64)

		rows = append(rows, row.New(6).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
			col.New(2).Add(text.New(det.CodigoPrincipal, props.Text{Size: 7, Align: align.Left, Top: 1, Left: 2})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", cant), props.Text{Size: 7, Align: align.Center, Top: 1})),
			col.New(5).Add(text.New(det.Descripcion, props.Text{Size: 7, Align: align.Left, Top: 1, Left: 2})),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", pUnit), props.Text{Size: 7, Align: align.Right, Top: 1, Right: 2})),
			col.New(1).Add(text.New(det.Descuento, props.Text{Size: 7, Align: align.Right, Top: 1, Right: 2})),
			col.New(1).Add(text.New(det.PrecioTotalSinImpuesto, props.Text{Size: 7, Align: align.Right, Top: 1, Right: 2})),
		))
	}
	return rows
}

func (g *RideGenerator) buildFooter(f *Factura) []core.Row {
	sub15, sub0, totalIVA := "0.00", "0.00", "0.00"
	for _, tax := range f.InfoFactura.TotalConImpuestos.TotalImpuesto {
		switch tax.CodigoPorcentaje {
		case "4":
			sub15 = tax.BaseImponible
			totalIVA = tax.Valor
		case "0":
			sub0 = tax.BaseImponible
		}
	}

	formaPago := "Sin Utilización del Sistema Financiero"
	if len(f.InfoFactura.Pagos.Pago) > 0 {
		formaPago = f.InfoFactura.Pagos.Pago[0].FormaPago
	}
	valorPago := f.InfoFactura.ImporteTotal

	// ALTURA TOTAL DE LA TABLA DE TOTALES = 14 filas * 5mm aprox = 70-75mm
	totalHeight := 72.0

	return []core.Row{
		row.New(totalHeight).Add(
			// Izquierda: Info Adicional y Pagos (Con borde completo)
			col.New(7).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
				text.New("Información Adicional", props.Text{Style: fontstyle.Bold, Size: 8, Top: 2, Left: 2}),
				text.New("Dirección: "+f.InfoFactura.DireccionComprador, props.Text{Size: 7, Top: 8, Left: 2}),
				text.New("Email: "+f.InfoFactura.DirEstablecimiento, props.Text{Size: 7, Top: 14, Left: 2}),

				// Tabla de Pagos Simulada
				text.New("Forma de Pago", props.Text{Style: fontstyle.Bold, Size: 7, Top: 30, Left: 2}),
				text.New("Valor", props.Text{Style: fontstyle.Bold, Size: 7, Top: 30, Align: align.Right, Right: 10}),

				text.New(formaPago, props.Text{Size: 6, Top: 35, Left: 2}),
				text.New(valorPago, props.Text{Size: 7, Top: 35, Align: align.Right, Right: 10}),
			),

			col.New(5).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
				// Fila 1: Subtotal 12%
				text.New("SUBTOTAL 12%", props.Text{Size: 7, Top: 2, Left: 2}),
				text.New(sub15, props.Text{Size: 7, Align: align.Right, Top: 2, Right: 2}),

				// Fila 2: Subtotal 0%
				text.New("SUBTOTAL 0%", props.Text{Size: 7, Top: 7, Left: 2}),
				text.New(sub0, props.Text{Size: 7, Align: align.Right, Top: 7, Right: 2}),

				// Fila 3: No Objeto
				text.New("SUBTOTAL NO OBJETO DE IVA", props.Text{Size: 7, Top: 12, Left: 2}),
				text.New("0.00", props.Text{Size: 7, Align: align.Right, Top: 12, Right: 2}),

				// Fila 4: Exento
				text.New("SUBTOTAL EXENTO DE IVA", props.Text{Size: 7, Top: 17, Left: 2}),
				text.New("0.00", props.Text{Size: 7, Align: align.Right, Top: 17, Right: 2}),

				// Fila 5: Sin Impuestos
				text.New("SUBTOTAL SIN IMPUESTOS", props.Text{Size: 7, Top: 22, Left: 2}),
				text.New(f.InfoFactura.TotalSinImpuestos, props.Text{Size: 7, Align: align.Right, Top: 22, Right: 2}),

				// Fila 6: Descuento
				text.New("TOTAL DESCUENTO", props.Text{Size: 7, Top: 27, Left: 2}),
				text.New(f.InfoFactura.TotalDescuento, props.Text{Size: 7, Align: align.Right, Top: 27, Right: 2}),

				// Fila 7: ICE
				text.New("ICE", props.Text{Size: 7, Top: 32, Left: 2}),
				text.New("0.00", props.Text{Size: 7, Align: align.Right, Top: 32, Right: 2}),

				// Fila 8: IVA
				text.New("IVA 12%", props.Text{Size: 7, Top: 37, Left: 2}),
				text.New(totalIVA, props.Text{Size: 7, Align: align.Right, Top: 37, Right: 2}),

				// Fila 9: Propina
				text.New("PROPINA", props.Text{Size: 7, Top: 42, Left: 2}),
				text.New("0.00", props.Text{Size: 7, Align: align.Right, Top: 42, Right: 2}),

				// Fila 10: TOTAL (Negrita)
				text.New("VALOR TOTAL", props.Text{Style: fontstyle.Bold, Size: 8, Top: 47, Left: 2}),
				text.New(f.InfoFactura.ImporteTotal, props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Right, Top: 47, Right: 2}),

				// Fila 11: Total sin Subsidio
				text.New("VALOR TOTAL SIN SUBSIDIO", props.Text{Size: 7, Top: 55, Left: 2}),
				text.New("0.00", props.Text{Size: 7, Align: align.Right, Top: 55, Right: 2}),

				// Fila 12: Ahorro Subsidio
				text.New("AHORRO POR SUBSIDIO:", props.Text{Size: 7, Top: 60, Left: 2}),
				text.New("0.00", props.Text{Size: 7, Align: align.Right, Top: 60, Right: 2}),

				// Fila 13: Nota Subsidio
				text.New("(Incluye IVA cuando corresponda)", props.Text{Size: 6, Top: 65, Left: 2}),
			),
		),
	}
}

func (g *RideGenerator) GenerateNotaCreditoRide(nc *NotaCredito, ridePath string, logoPath string, authDate time.Time, authNumber string) error {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithOrientation(orientation.Vertical).
		WithLeftMargin(10).
		WithRightMargin(10).
		WithTopMargin(10).
		Build()

	m := maroto.New(cfg)

	m.AddRows(g.buildHeaderNC(nc, logoPath, authDate, authNumber)...)
	m.AddRow(2)
	m.AddRows(g.buildClientInfoNC(nc)...)
	m.AddRow(2)
	m.AddRows(g.buildDetailsRowsNC(nc)...)
	m.AddRow(2)
	m.AddRows(g.buildFooterNC(nc)...)

	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("error generando RIDE NC: %w", err)
	}

	return document.Save(ridePath)
}

func (g *RideGenerator) buildHeaderNC(nc *NotaCredito, logoPath string, authDate time.Time, authNumber string) []core.Row {
	ambiente := "PRUEBAS"
	if nc.InfoTributaria.Ambiente == "2" {
		ambiente = "PRODUCCIÓN"
	}

	var logo core.Component
	if logoPath != "" {
		logo = image.NewFromFile(logoPath)
	} else {
		logo = text.New("NO TIENE LOGO", props.Text{
			Style: fontstyle.Bold, Size: 16, Align: align.Center, Top: 10, Color: &props.Color{Red: 200},
		})
	}

	return []core.Row{
		row.New(90).Add(
			col.New(6).Add(
				logo,
				text.New(" ", props.Text{Top: 35}), // Spacer
				text.New(nc.InfoTributaria.RazonSocial, props.Text{Style: fontstyle.Bold, Size: 8, Top: 40}),
				text.New(nc.InfoTributaria.NombreComercial, props.Text{Size: 8, Top: 45}),
				text.New("Dirección Matriz:", props.Text{Style: fontstyle.Bold, Size: 8, Top: 55}),
				text.New(nc.InfoTributaria.DirMatriz, props.Text{Size: 8, Top: 60}),
				text.New("Dirección Sucursal:", props.Text{Style: fontstyle.Bold, Size: 8, Top: 68}),
				text.New(nc.InfoNotaCredito.DirEstablecimiento, props.Text{Size: 8, Top: 73}),
				text.New("OBLIGADO A LLEVAR CONTABILIDAD: "+nc.InfoNotaCredito.ObligadoContabilidad, props.Text{Size: 8, Top: 82}),
			),
			col.New(6).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
				text.New("R.U.C.: "+nc.InfoTributaria.Ruc, props.Text{Style: fontstyle.Bold, Size: 10, Top: 3, Left: 2}),
				text.New("NOTA DE CRÉDITO", props.Text{Style: fontstyle.Bold, Size: 12, Top: 9, Left: 2}),
				text.New("No. "+fmt.Sprintf("%s-%s-%s", nc.InfoTributaria.Estab, nc.InfoTributaria.PtoEmi, nc.InfoTributaria.Secuencial), props.Text{Size: 10, Top: 15, Left: 2}),
				text.New("NÚMERO DE AUTORIZACIÓN", props.Text{Style: fontstyle.Bold, Size: 8, Top: 22, Left: 2}),
				text.New(authNumber, props.Text{Size: 8, Top: 26, Left: 2}),
				text.New("FECHA Y HORA DE AUTORIZACIÓN", props.Text{Style: fontstyle.Bold, Size: 8, Top: 32, Left: 2}),
				text.New(authDate.Format("02/01/2006 15:04:05"), props.Text{Size: 8, Top: 36, Left: 2}),
				text.New("AMBIENTE: "+ambiente, props.Text{Size: 8, Top: 42, Left: 2}),
				text.New("EMISIÓN: NORMAL", props.Text{Size: 8, Top: 48, Left: 2}),
				text.New("CLAVE DE ACCESO", props.Text{Style: fontstyle.Bold, Size: 8, Top: 60, Left: 2}),
				code.NewBar(nc.InfoTributaria.ClaveAcceso, props.Barcode{Percent: 67, Top: 68, Left: 16}),
				text.New(nc.InfoTributaria.ClaveAcceso, props.Text{Size: 7, Align: align.Center, Top: 82}),
			),
		),
	}
}

func (g *RideGenerator) buildClientInfoNC(nc *NotaCredito) []core.Row {
	return []core.Row{
		row.New(35).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
			col.New(7).Add(
				text.New("Razón Social / Nombres y Apellidos: "+nc.InfoNotaCredito.RazonSocialComprador, props.Text{Style: fontstyle.Bold, Size: 7, Top: 2, Left: 2}),
				text.New("Identificación: "+nc.InfoNotaCredito.IdentificacionComprador, props.Text{Size: 7, Top: 8, Left: 2}),
				text.New("Fecha Emisión: "+nc.InfoNotaCredito.FechaEmision, props.Text{Size: 7, Top: 14, Left: 2}),
				text.New("------------------------------------------------------------------------------------------------------", props.Text{Size: 5, Top: 19}),
				text.New("Comprobante que se modifica:", props.Text{Style: fontstyle.Bold, Size: 7, Top: 22, Left: 2}),
				text.New("FACTURA "+nc.InfoNotaCredito.NumDocModificado, props.Text{Size: 7, Top: 22, Left: 45}),
				text.New("Fecha Emisión (Comprobante a modificar):", props.Text{Style: fontstyle.Bold, Size: 7, Top: 27, Left: 2}),
				text.New(nc.InfoNotaCredito.FechaEmisionDocSustento, props.Text{Size: 7, Top: 27, Left: 60}),
				text.New("Razón de Modificación:", props.Text{Style: fontstyle.Bold, Size: 7, Top: 32, Left: 2}),
				text.New(nc.InfoNotaCredito.Motivo, props.Text{Size: 7, Top: 32, Left: 35}),
			),
			col.New(5).Add(
				text.New("R.U.C. / C.I.: "+nc.InfoNotaCredito.IdentificacionComprador, props.Text{Size: 7, Top: 8, Align: align.Right, Right: 5}),
			),
		),
	}
}

func (g *RideGenerator) buildDetailsRowsNC(nc *NotaCredito) []core.Row {
	var rows []core.Row
	headerStyle := &props.Cell{BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240}, BorderType: border.Full, BorderThickness: 0.1}

	rows = append(rows, row.New(8).WithStyle(headerStyle).Add(
		col.New(2).Add(text.New("Cod. Interno", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Center, Top: 1.5})),
		col.New(1).Add(text.New("Cant.", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Center, Top: 1.5})),
		col.New(5).Add(text.New("Descripción", props.Text{Style: fontstyle.Bold, Size: 7, Top: 1.5, Left: 2})),
		col.New(2).Add(text.New("Precio Unitario", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Right, Top: 1.5, Right: 2})),
		col.New(1).Add(text.New("Desc.", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Right, Top: 1.5, Right: 2})),
		col.New(1).Add(text.New("Precio Total", props.Text{Style: fontstyle.Bold, Size: 7, Align: align.Right, Top: 1.5, Right: 2})),
	))

	for _, det := range nc.Detalles.Detalle {
		cant, _ := strconv.ParseFloat(det.Cantidad, 64)
		pUnit, _ := strconv.ParseFloat(det.PrecioUnitario, 64)

		rows = append(rows, row.New(6).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
			col.New(2).Add(text.New(det.CodigoInterno, props.Text{Size: 7, Align: align.Left, Top: 1, Left: 2})),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", cant), props.Text{Size: 7, Align: align.Center, Top: 1})),
			col.New(5).Add(text.New(det.Descripcion, props.Text{Size: 7, Align: align.Left, Top: 1, Left: 2})),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", pUnit), props.Text{Size: 7, Align: align.Right, Top: 1, Right: 2})),
			col.New(1).Add(text.New(det.Descuento, props.Text{Size: 7, Align: align.Right, Top: 1, Right: 2})),
			col.New(1).Add(text.New(det.PrecioTotalSinImpuesto, props.Text{Size: 7, Align: align.Right, Top: 1, Right: 2})),
		))
	}
	return rows
}

func (g *RideGenerator) buildFooterNC(nc *NotaCredito) []core.Row {
	sub15, sub0, totalIVA := "0.00", "0.00", "0.00"
	for _, tax := range nc.InfoNotaCredito.TotalConImpuestos.TotalImpuesto {
		switch tax.CodigoPorcentaje {
		case "4":
			sub15 = tax.BaseImponible
			totalIVA = tax.Valor
		case "0":
			sub0 = tax.BaseImponible
		}
	}

	totalHeight := 60.0
	return []core.Row{
		row.New(totalHeight).Add(
			col.New(7).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
				text.New("Información Adicional", props.Text{Style: fontstyle.Bold, Size: 8, Top: 2, Left: 2}),
				text.New("Nota de Crédito generada automáticamente.", props.Text{Size: 7, Top: 8, Left: 2}),
			),
			col.New(5).WithStyle(&props.Cell{BorderType: border.Full, BorderThickness: 0.1}).Add(
				text.New("SUBTOTAL 12%", props.Text{Size: 7, Top: 2, Left: 2}),
				text.New(sub15, props.Text{Size: 7, Align: align.Right, Top: 2, Right: 2}),
				text.New("SUBTOTAL 0%", props.Text{Size: 7, Top: 7, Left: 2}),
				text.New(sub0, props.Text{Size: 7, Align: align.Right, Top: 7, Right: 2}),
				text.New("SUBTOTAL SIN IMPUESTOS", props.Text{Size: 7, Top: 12, Left: 2}),
				text.New(nc.InfoNotaCredito.TotalSinImpuestos, props.Text{Size: 7, Align: align.Right, Top: 12, Right: 2}),
				text.New("IVA 12%", props.Text{Size: 7, Top: 17, Left: 2}),
				text.New(totalIVA, props.Text{Size: 7, Align: align.Right, Top: 17, Right: 2}),
				text.New("VALOR TOTAL", props.Text{Style: fontstyle.Bold, Size: 8, Top: 22, Left: 2}),
				text.New(nc.InfoNotaCredito.ValorModificacion, props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Right, Top: 22, Right: 2}),
			),
		),
	}
}
