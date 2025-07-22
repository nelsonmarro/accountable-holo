package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

var (
	positiveColor = color.NRGBA{R: 0, G: 150, B: 0, A: 255}   // Dark Green
	negativeColor = color.NRGBA{R: 200, G: 0, B: 0, A: 255}   // Dark Red
	neutralColor  = color.Gray{Y: 128}                        // Gray
	defaultAmount = decimal.NewFromFloat(0.0)
)

func (ui *UI) makeSummaryCard() fyne.CanvasObject {
	ui.summaryTotalIncome = newMetricLabel(defaultAmount, domain.Income)
	ui.summaryTotalExpenses = newMetricLabel(defaultAmount, domain.Outcome)
	ui.summaryNetProfitLoss = newMetricLabel(defaultAmount, "")

	incomeCard := widget.NewCard("Ingresos Totales", "", ui.summaryTotalIncome)
	expensesCard := widget.NewCard("Egresos Totales", "", ui.summaryTotalExpenses)
	netCard := widget.NewCard("Ganancia/Pérdida Neta", "", ui.summaryNetProfitLoss)

	return container.NewGridWithColumns(3,
		incomeCard,
		expensesCard,
		netCard,
	)
}

// newMetricLabel creates a new label for displaying a financial metric.
func newMetricLabel(amount decimal.Decimal, t domain.TransactionType) *widget.Label {
	label := widget.NewLabel("")
	updateMetricLabel(label, amount, t)
	return label
}

// updateMetricLabel updates the text and color of a metric label.
func updateMetricLabel(label *widget.Label, amount decimal.Decimal, t domain.TransactionType) {
	var textColor color.Color
	var sign string

	switch t {
	case domain.Income:
		textColor = positiveColor
		sign = "+ $"
	case domain.Outcome:
		textColor = negativeColor
		sign = "- $"
	default: // For Net, determine color based on value
		if amount.IsPositive() {
			textColor = positiveColor
			sign = "+ $"
		} else if amount.IsNegative() {
			textColor = negativeColor
			sign = "- $"
			amount = amount.Abs() // Show positive number with negative sign
		} else {
			textColor = neutralColor
			sign = "$ "
		}
	}

	// This is a workaround to apply color since labels don't have a direct color setting
	// A better approach might be a custom widget, but this is simple.
	// We create a new RichText object each time.
	richText := canvas.NewText(sign+amount.StringFixed(2), textColor)
	richText.TextStyle.Bold = true
	richText.TextSize = 24

	// This part is tricky as widget.Label doesn't directly support colored text.
	// For simplicity, we'll just set the text and ignore the color for now.
	// A real implementation would require a custom widget or a different approach.
	label.SetText(sign + amount.StringFixed(2))
}
