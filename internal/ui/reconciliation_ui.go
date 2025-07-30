package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeReconciliationUI() fyne.CanvasObject {
	// UI elelennts
	title := widget.NewLabel("Reconciliación de Cuentas")

	// Top card (form)
	ui.makeFormCard()

	// Bottom card (reconciliation statement)
	ui.makeStatementCard()

	// containers
	return widget.NewLabel("Reconciliation UI is under construction. Please check back later.")
}

func (ui *UI) makeFormCard() fyne.CanvasObject {
	accountsSelector := widget.NewSelect([]string{}, nil) // we'll populate this later
	endingDateEntry := widget.NewDateEntry()
	actualBalanceEntry := widget.NewEntry()
}

func (ui *UI) makeStatementCard() fyne.CanvasObject {
	return widget.NewLabel("Statement Card is under construction. Please check back later.")
}
