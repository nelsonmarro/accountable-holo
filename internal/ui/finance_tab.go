package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeFinanceTab() fyne.CanvasObject {
	// UI widgets

	// Containers
	transactionsTab := container.NewTabItem("Transacciones", widget.NewLabel("Transacciones"))
	categoriesTab := container.NewTabItem("Categorías", widget.NewLabel("Categorías"))

	tabContainer := container.NewAppTabs(transactionsTab, categoriesTab)
	tabContainer.SetTabLocation(container.TabLocationTrailing)

	return tabContainer
}
