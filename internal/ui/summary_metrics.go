package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeSummaryCard() fyne.CanvasObject {
	return widget.NewLabel("Metrics")
}
