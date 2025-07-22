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

	// Containers
	mainLayout := container.NewVBox(
		container.NewPadded(filterCard),
		container.NewPadded(summaryCard),
	)

	return mainLayout
}
