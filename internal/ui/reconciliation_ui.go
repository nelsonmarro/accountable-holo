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

	// TODO: add validation for the balance entry

	reconciliationForm := widget.NewForm(
		widget.NewFormItem("Cuenta", accountsSelector),
		widget.NewFormItem("Fecha de cierre", endingDateEntry),
		widget.NewFormItem("Saldo Final Real", actualBalanceEntry),
	)

	reconciliationForm.OnSubmit = func() {
		// TODO: Implement the reconciliation logic
		// a. Parse the values from the form widgets.
		// b. Call the ui.Services.TxService.ReconcileAccount method.
		// c. Take the result and populate the statement card.
		// d. Show the statement card.
	}

	backButton := widget.NewButton("Volver", func() {
		// This should navigate back to the main transaction view.
		// You can call the navigation function you created earlier.
		ui.navToView(ui.makeFinancesTab())
	})

	// Don't forget to load the accounts for the selector, similar to how you do it in the
	go ui.loadAccountsForReconciliation(accountSelector)

	return nil
}

func (ui *UI) makeStatementCard() fyne.CanvasObject {
	return widget.NewLabel("Statement Card is under construction. Please check back later.")
}
