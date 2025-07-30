package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

func (ui *UI) makeReconciliationUI() fyne.CanvasObject {
	// Top card (form)

	// Bottom card (reconciliation statement)

	// containers
	mainContainer := container.NewBorder(
		container.NewPadded(ui.makeFormCard()),
		nil, nil, nil,
		container.NewPadded(ui.makeStatementCard()),
	)
	return mainContainer
}

func (ui *UI) makeFormCard() fyne.CanvasObject {
	accountsSelector := widget.NewSelectEntry([]string{}) // we'll populate this later
	endingDateEntry := widget.NewDateEntry()
	actualBalanceEntry := widget.NewEntry()

	// TODO: add validation for the balance entry
	formValidation(accountsSelector, endingDateEntry, actualBalanceEntry)

	reconciliationForm := widget.NewForm(
		widget.NewFormItem("Cuenta", accountsSelector),
		widget.NewFormItem("Fecha de cierre", endingDateEntry),
		widget.NewFormItem("Saldo Final Real", actualBalanceEntry),
	)

	reconciliationForm.OnSubmit = func() {
		selectedAccountName := accountsSelector.Text
	}

	backButton := widget.NewButton("Volver", func() {
		// This should navigate back to the main transaction view.
		// You can call the navigation function you created earlier.
		ui.navToView(ui.makeFinancesTab())
	})

	// Create the card itself
	formCard := widget.NewCard(
		"Reconciliación de Cuenta",
		"",
		container.NewVBox(reconciliationForm, backButton),
	)

	// Don't forget to load the accounts for the selector, similar to how you do it in the
	go ui.loadAccountsForReconciliation(accountsSelector)

	return formCard
}

func (ui *UI) makeStatementCard() fyne.CanvasObject {
	return widget.NewLabel("Statement Card is under construction. Please check back later.")
}

func (ui *UI) loadAccountsForReconciliation(selector *widget.SelectEntry) {
	var accounts []domain.Account

	if ui.accounts == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		accs, err := ui.Services.AccService.GetAllAccounts(ctx)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("error al cargar las cuentas: %v", err), ui.mainWindow)
			})
			return
		}
		accounts = accs
	}

	accountNames := make([]string, len(accounts))
	for i, acc := range accounts {
		accountNames[i] = acc.Name
	}

	selector.SetOptions(accountNames)
}

func formValidation(
	accountsSelector *widget.SelectEntry,
	endingDateEntry *widget.DateEntry,
	actualBalanceEntry *widget.Entry,
) {
	// Add validation logic here
	// For example, you can disable the submit button until all fields are valid
	// or show error messages if the fields are not filled correctly.
	// This is a placeholder for your validation logic.
}
