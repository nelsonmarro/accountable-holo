package ui

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

var (
	positiveColor = color.NRGBA{R: 0, G: 150, B: 0, A: 255} // Dark Green
	negativeColor = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Dark Red
	neutralColor  = color.Gray{Y: 128}                      // Gray
	defaultAmount = decimal.NewFromFloat(0.0)
)

// makeSummaryCard creates a card with financial metrics for the summary dashboard.
func (ui *UI) makeSummaryCard() fyne.CanvasObject {
	ui.summaryTotalIncome = newMetricText(defaultAmount, domain.Income)
	ui.summaryTotalExpenses = newMetricText(defaultAmount, domain.Outcome)
	ui.summaryNetProfitLoss = newMetricText(defaultAmount, "Net") // Use a specific type for Net

	incomeCard := widget.NewCard("Ingresos Totales", "", container.NewCenter(ui.summaryTotalIncome))
	expensesCard := widget.NewCard("Egresos Totales", "", container.NewCenter(ui.summaryTotalExpenses))
	netCard := widget.NewCard("Ganancia/Pérdida Neta", "", container.NewCenter(ui.summaryNetProfitLoss))

	return container.NewGridWithColumns(3,
		incomeCard,
		expensesCard,
		netCard,
	)
}

// newMetricText creates a new canvas.Text for displaying a financial metric.
func newMetricText(amount decimal.Decimal, t domain.CategoryType) *canvas.Text {
	text := &canvas.Text{
		Text:     "",
		TextSize: 24,
		TextStyle: fyne.TextStyle{
			Bold: true,
		},
		Alignment: fyne.TextAlignCenter,
	}
	updateMetricText(text, amount, t)
	return text
}

// updateMetricText updates the text and color of a metric canvas.Text object.
func updateMetricText(text *canvas.Text, amount decimal.Decimal, t domain.CategoryType) {
	var sign string

	switch t {
	case domain.Income:
		text.Color = positiveColor
		sign = "+ $"
	case domain.Outcome:
		text.Color = negativeColor
		sign = "- $"
	default: // For Net, determine color based on value
		if amount.IsPositive() {
			text.Color = positiveColor
			sign = "+ $"
		} else if amount.IsNegative() {
			text.Color = negativeColor
			sign = "- $"
			amount = amount.Abs() // Show positive number with negative sign
		} else {
			text.Color = neutralColor
			sign = "$ "
		}
	}

	text.Text = fmt.Sprintf("%s%s", sign, amount.StringFixed(2))
	text.Refresh()
}
