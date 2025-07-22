package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeFilterCard() fyne.CanvasObject {
	return widget.NewLabel("Filters")
}
