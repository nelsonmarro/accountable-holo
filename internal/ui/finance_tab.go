package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func (ui *UI) makeTransactionsTab() fyne.CanvasObject {
	// UI widgets
	transactionsTab := container.NewTabItem("Transacciones", ui.makeTransactionUI())
	categoriesTab := container.NewTabItem("Categorías", ui.makeCategoryUI())

	// Containers
	tabContainer := container.NewAppTabs(transactionsTab, categoriesTab)
	tabContainer.SetTabLocation(container.TabLocationBottom)

	return tabContainer
}
