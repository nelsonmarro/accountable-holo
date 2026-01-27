package componets

import (
	"fmt"
	"image/color"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/vicanso/go-charts/v2"
)

// RenderIncomeExpenseChart genera un gráfico de barras comparativo.
func RenderIncomeExpenseChart(income, expense decimal.Decimal) fyne.CanvasObject {
	incFloat, _ := income.Float64()
	expFloat, _ := expense.Float64()

	if incFloat == 0 && expFloat == 0 {
		return container.NewCenter(widget.NewLabel("Sin datos financieros"))
	}

	// Definir colores personalizados compatibles con go-charts (drawing.Color)
	green := color.NRGBA{R: 34, G: 197, B: 94, A: 255}
	red := color.NRGBA{R: 239, G: 68, B: 68, A: 255}

	p, err := charts.BarRender(
		[][]float64{{incFloat}, {expFloat}},
		charts.LegendOptionFunc(charts.LegendOption{
			Data: []string{"Ingresos", "Egresos"},
			Left: charts.PositionCenter,
			Top:  charts.PositionBottom,
		}),
		charts.XAxisDataOptionFunc([]string{"Periodo Seleccionado"}),
		charts.ThemeOptionFunc("grafiti"),
		// Aplicamos colores manualmente a través de una función anónima (OptionFunc)
		func(opt *charts.ChartOption) {
			if len(opt.SeriesList) >= 2 {
				opt.SeriesList[0].Style.FillColor = charts.Color{R: green.R, G: green.G, B: green.B, A: green.A}
				opt.SeriesList[1].Style.FillColor = charts.Color{R: red.R, G: red.G, B: red.B, A: red.A}
				// También el borde para que se vea consistente
				opt.SeriesList[0].Style.StrokeColor = opt.SeriesList[0].Style.FillColor
				opt.SeriesList[1].Style.StrokeColor = opt.SeriesList[1].Style.FillColor
			}
		},
		charts.WidthOptionFunc(900), // Resolución optimizada para legibilidad de texto
		charts.HeightOptionFunc(550),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  20,
			Bottom: 50,
			Left:   50,
		}),
	)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	buf, err := p.Bytes()
	if err != nil {
		return widget.NewLabel("Error renderizando: " + err.Error())
	}

	res := fyne.NewStaticResource("bar_chart.png", buf)
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(700, 400)) // Más grande en pantalla

	return img
}

// RenderCategoryPieChart genera un gráfico de donut para distribución de gastos.
func RenderCategoryPieChart(data []domain.CategoryAmount) fyne.CanvasObject {
	if len(data) == 0 {
		return container.NewCenter(widget.NewLabel("Sin datos de gastos"))
	}

	// Ordenar y Limitar
	sort.Slice(data, func(i, j int) bool {
		return data[i].Amount.GreaterThan(data[j].Amount)
	})

	var values []float64
	var labels []string
	var otherTotal decimal.Decimal
	limit := 5

	// Calcular total para porcentajes manuales
	var totalAmount decimal.Decimal
	for _, item := range data {
		totalAmount = totalAmount.Add(item.Amount)
	}

	for i, item := range data {
		if i < limit {
			val, _ := item.Amount.Float64()
			values = append(values, val)

			// Porcentaje manual para el label
			percent := item.Amount.Div(totalAmount).Mul(decimal.NewFromInt(100)).StringFixed(1)
			labels = append(labels, fmt.Sprintf("%s (%s%%)", item.CategoryName, percent))
		} else {
			otherTotal = otherTotal.Add(item.Amount)
		}
	}

	if otherTotal.IsPositive() {
		val, _ := otherTotal.Float64()
		values = append(values, val)
		percent := otherTotal.Div(totalAmount).Mul(decimal.NewFromInt(100)).StringFixed(1)
		labels = append(labels, fmt.Sprintf("Otros (%s%%)", percent))
	}

	p, err := charts.PieRender(
		values,
		charts.LegendOptionFunc(charts.LegendOption{
			Orient: charts.OrientHorizontal, // Horizontal para abajo
			Left:   charts.PositionCenter,
			Top:    charts.PositionBottom,
			Data:   labels,
		}),
		charts.PieSeriesShowLabel(),
		charts.ThemeOptionFunc("grafiti"),
		charts.WidthOptionFunc(900), // Resolución optimizada para legibilidad de texto
		charts.HeightOptionFunc(550),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  20,
			Bottom: 20,
			Left:   20,
		}),
	)
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	buf, err := p.Bytes()
	if err != nil {
		return widget.NewLabel("Error renderizando: " + err.Error())
	}

	res := fyne.NewStaticResource("pie_chart.png", buf)
	img := canvas.NewImageFromResource(res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(700, 400)) // Más grande en pantalla

	return img
}

// --- BUDGET COMPONENT (Sin cambios) ---
// (Mantenemos el código anterior del NewBudgetStatusChart que ya funcionaba bien y era nativo)

type budgetBarLayout struct {
	ratio float32
}

func (l *budgetBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	objects[0].Resize(size)
	objects[0].Move(fyne.NewPos(0, 0))

	fgWidth := size.Width * l.ratio
	objects[1].Resize(fyne.NewSize(fgWidth, size.Height))
	objects[1].Move(fyne.NewPos(0, 0))
}

func (l *budgetBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 12)
}

func NewBudgetStatusChart(statuses []domain.BudgetStatus, onSeeMore func()) fyne.CanvasObject {
	if len(statuses) == 0 {
		return widget.NewLabel("No hay presupuestos definidos para este periodo.")
	}

	limit := 5
	displayData := statuses
	showButton := false

	if onSeeMore != nil && len(statuses) > limit {
		displayData = statuses[:limit]
		showButton = true
	}

	containerList := container.NewVBox()

	for _, status := range displayData {
		progressVal := float32(status.PercentageUsed / 100.0)
		if progressVal > 1.0 {
			progressVal = 1.0
		}

		barColor := color.NRGBA{R: 34, G: 197, B: 94, A: 255}
		trackColor := color.NRGBA{R: 38, G: 38, B: 38, A: 255}

		if status.IsOverBudget {
			barColor = color.NRGBA{R: 220, G: 38, B: 38, A: 255}
		}

		nameLabel := widget.NewLabel(status.CategoryName)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		statusText := fmt.Sprintf("$%s / $%s (%.1f%%)",
			status.SpentAmount.StringFixed(0),
			status.BudgetAmount.StringFixed(0),
			status.PercentageUsed)

		infoLabel := widget.NewLabel(statusText)
		infoLabel.Alignment = fyne.TextAlignTrailing

		bgBar := canvas.NewRectangle(trackColor)
		bgBar.CornerRadius = 6
		fgBar := canvas.NewRectangle(barColor)
		fgBar.CornerRadius = 6

		barLayout := container.New(&budgetBarLayout{ratio: progressVal}, bgBar, fgBar)

		rowContent := container.NewVBox(
			container.NewBorder(nil, nil, nameLabel, infoLabel),
			barLayout,
		)

		if status.IsOverBudget {
			warningText := canvas.NewText("🚨 ¡PRESUPUESTO EXCEDIDO!", color.NRGBA{R: 239, G: 68, B: 68, A: 255})
			warningText.TextSize = 18
			warningText.TextStyle = fyne.TextStyle{Bold: true}
			warningText.Alignment = fyne.TextAlignTrailing
			rowContent.Add(container.NewPadded(warningText))
		}

		containerList.Add(container.NewPadded(rowContent))
	}

	if showButton {
		btn := widget.NewButton("Ver Más...", onSeeMore)
		btn.Importance = widget.LowImportance
		containerList.Add(container.NewPadded(btn))
	}

	return containerList
}
