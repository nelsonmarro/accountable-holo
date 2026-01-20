package componets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/shopspring/decimal"
)

// NewIncomeExpenseChart creates a simple vertical bar chart comparing Income vs Expenses.
func NewIncomeExpenseChart(income, expense decimal.Decimal) fyne.CanvasObject {
	maxVal := income
	if expense.GreaterThan(income) {
		maxVal = expense
	}

	if maxVal.IsZero() {
		return widget.NewLabel("Sin datos para mostrar")
	}

	// Helper to calculate height factor (0.0 to 1.0)
	getFactor := func(val decimal.Decimal) float32 {
		f, _ := val.Div(maxVal).Float64()
		return float32(f)
	}

	incomeFactor := getFactor(income)
	expenseFactor := getFactor(expense)

	// Let's use a fixed height container for the chart area, e.g., 200px.
	chartHeight := float32(200)

	createAlignedBar := func(_ decimal.Decimal, factor float32, col color.Color) fyne.CanvasObject {
		barPixelHeight := chartHeight * factor
		if barPixelHeight < 2 {
			barPixelHeight = 2
		}

		rect := canvas.NewRectangle(col)
		rect.SetMinSize(fyne.NewSize(40, barPixelHeight))

		// Invisible spacer to take up the rest of the space
		spacerHeight := chartHeight - barPixelHeight
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(40, spacerHeight))

		return container.NewVBox(spacer, rect)
	}

	incomeCol := container.NewVBox(
		widget.NewLabelWithStyle(fmt.Sprintf("$%s", income.StringFixed(0)), fyne.TextAlignCenter, fyne.TextStyle{}),
		createAlignedBar(income, incomeFactor, color.NRGBA{R: 0, G: 150, B: 0, A: 255}),
		widget.NewLabelWithStyle("Ingresos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	expenseCol := container.NewVBox(
		widget.NewLabelWithStyle(fmt.Sprintf("$%s", expense.StringFixed(0)), fyne.TextAlignCenter, fyne.TextStyle{}),
		createAlignedBar(expense, expenseFactor, color.NRGBA{R: 200, G: 0, B: 0, A: 255}),
		widget.NewLabelWithStyle("Egresos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	return container.NewGridWithColumns(2, incomeCol, expenseCol)
}

// NewCategoryBreakdownChart creates a horizontal bar chart for top categories with optional 'Ver Más' button.
func NewCategoryBreakdownChart(data []domain.CategoryAmount, total decimal.Decimal, onSeeMore func()) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("Sin datos de gastos")
	}

	limit := 5
	displayData := data
	showButton := false

	if onSeeMore != nil && len(data) > limit {
		displayData = data[:limit]
		showButton = true
	}

	// Container for rows
	rows := container.NewVBox()

	for _, item := range displayData {
		percent := item.Amount.Div(total).Mul(decimal.NewFromFloat(100))
		percentFloat, _ := percent.Float64()

		// Label: "Category (25%)"
		label := widget.NewLabel(fmt.Sprintf("%s (%.1f%%)", item.CategoryName, percentFloat))
		label.TextStyle = fyne.TextStyle{Bold: true}

		// Bar (Progress Bar is perfect for horizontal bars!)
		progressBar := widget.NewProgressBar()
		progressBar.Min = 0
		totalFloat, _ := total.Float64()
		progressBar.Max = totalFloat
		valFloat, _ := item.Amount.Float64()
		progressBar.SetValue(valFloat)
		progressBar.TextFormatter = func() string { return "" } // Hide percentage text inside bar

		// Amount Label
		amountLabel := widget.NewLabel(fmt.Sprintf("$%s", item.Amount.StringFixed(2)))
		amountLabel.Alignment = fyne.TextAlignTrailing

		// Row Layout
		header := container.NewBorder(nil, nil, label, amountLabel)
		rows.Add(container.NewVBox(header, progressBar))
	}

	if showButton {
		btn := widget.NewButton("Ver Más...", onSeeMore)
		btn.Importance = widget.LowImportance
		rows.Add(container.NewPadded(btn))
	}

	return rows
}

// budgetBarLayout es un layout personalizado para que la barra de progreso sea responsiva
type budgetBarLayout struct {
	ratio float32
}

func (l *budgetBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	// El fondo (bg) ocupa todo el espacio
	objects[0].Resize(size)
	objects[0].Move(fyne.NewPos(0, 0))

	// El frente (fg) ocupa solo el porcentaje indicado
	fgWidth := size.Width * l.ratio
	objects[1].Resize(fyne.NewSize(fgWidth, size.Height))
	objects[1].Move(fyne.NewPos(0, 0))
}

func (l *budgetBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 12) // Tamaño mínimo razonable
}

// NewBudgetStatusChart crea una lista visual de estados de presupuesto con barras responsivas y botón 'Ver Más' opcional.
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

		// Colores mejorados
		barColor := color.NRGBA{R: 34, G: 197, B: 94, A: 255}  // Verde esmeralda
		trackColor := color.NRGBA{R: 38, G: 38, B: 38, A: 255} // Gris carbón muy oscuro

		if status.IsOverBudget {
			barColor = color.NRGBA{R: 220, G: 38, B: 38, A: 255} // Rojo vibrante
		}

		// Etiqueta de Categoría
		nameLabel := widget.NewLabel(status.CategoryName)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		// Texto informativo: monto y porcentaje
		statusText := fmt.Sprintf("$%s / $%s (%.1f%%)",
			status.SpentAmount.StringFixed(0),
			status.BudgetAmount.StringFixed(0),
			status.PercentageUsed)

		infoLabel := widget.NewLabel(statusText)
		infoLabel.Alignment = fyne.TextAlignTrailing

		// Construcción de la barra responsiva
		bgBar := canvas.NewRectangle(trackColor)
		bgBar.CornerRadius = 6

		fgBar := canvas.NewRectangle(barColor)
		fgBar.CornerRadius = 6

		// Usamos nuestro layout personalizado
		barLayout := container.New(&budgetBarLayout{ratio: progressVal}, bgBar, fgBar)

		rowContent := container.NewVBox(
			container.NewBorder(nil, nil, nameLabel, infoLabel),
			barLayout,
		)

		if status.IsOverBudget {
			// Alerta más grande y con mejor espaciado
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
