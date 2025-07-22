package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func (ui *UI) makeSummaryTab() fyne.CanvasObject {
	// UI widgets
	filterCard := ui.makeFilterCard()
	summaryCard := ui.makeSummaryCard()

	// Containers
	mainLayout := container.NewVBox(
		filterCard,
		summaryCard,
	)

	return mainLayout
}
