package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeSummaryTab() fyne.CanvasObject {
	// UI widgets
	filterCard := widget.NewCard("Filtros", "", ui.makeFilterCard())
	summaryCard := widget.NewCard("Metricas", "", ui.makeSummaryCard())

	ui.summaryChartsContainer = container.NewGridWithColumns(2)
	chartsCard := widget.NewCard("Tendencias", "", ui.summaryChartsContainer)

	ui.summaryBudgetContainer = container.NewVBox()
	budgetCard := widget.NewCard("Control de Presupuestos", "", ui.summaryBudgetContainer)

	// Containers
	mainLayout := container.NewVBox(
		container.NewPadded(filterCard),
		container.NewPadded(summaryCard),
		container.NewPadded(chartsCard),
		container.NewPadded(budgetCard),
	)

	return container.NewVScroll(mainLayout)
}
