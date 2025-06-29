package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeFinanceTab() fyne.CanvasObject {
	// UI widgets
	transactionsTab := container.NewTabItem("Transacciones", widget.NewLabel("Transacciones"))
	categoriesTab := container.NewTabItem("Categorías", widget.NewLabel("Categorías"))

	// Containers
	tabContainer := container.NewAppTabs(transactionsTab, categoriesTab)
	tabContainer.SetTabLocation(container.TabLocationBottom)

	return tabContainer
}
