package transaction

import (
	"context"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/transaction"
	"github.com/shopspring/decimal"
)

type ReconciliationDialog struct {
	ui *ui.UI
	dialog.Dialog

	statementUI fyne.CanvasObject
	data        *domain.Reconciliation
	widgets     *reconciliationUIWidgets
}

func (ui *UI) makeFormCard() fyne.CanvasObject {
	accountsSelector := widget.NewSelectEntry([]string{}) // we'll populate this later
	endingDateEntry := widget.NewDateEntry()
	actualBalanceEntry := widget.NewEntry()

	// Validations for the balance entry
	formValidation(accountsSelector, endingDateEntry, actualBalanceEntry)

	reconciliationForm := widget.NewForm(
		widget.NewFormItem("Cuenta", accountsSelector),
		widget.NewFormItem("Fecha de cierre", endingDateEntry),
		widget.NewFormItem("Saldo Final Real", actualBalanceEntry),
	)

	reconciliationForm.OnSubmit = func() {
		selectedAccountName := accountsSelector.Text
		var selectedAccountID int
		for _, acc := range ui.accounts {
			if acc.Name == selectedAccountName {
				selectedAccountID = acc.ID
				break
			}
		}

		if selectedAccountID == 0 {
			dialog.ShowError(fmt.Errorf("la cuenta seleccionada no es válida"), ui.mainWindow)
			return
		}

		endingDate := endingDateEntry.Date

		actualBalance, err := decimal.NewFromString(actualBalanceEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("saldo final real no es un número válido: %v", err), ui.mainWindow)
			return
		}

		go ui.initiateReconciliation(selectedAccountID, endingDate, actualBalance)
	}

	backButton := widget.NewButton("Volver", func() {
		// This should navigate back to the main transaction view.
		// You can call the navigation function you created earlier.
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

func (ui *UI) initiateReconciliation(accountID int, endingDate *time.Time, actualBalance decimal.Decimal) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reconciliation, err := ui.Services.TxService.ReconcileAccount(ctx, accountID, *endingDate, actualBalance)
	if err != nil {
		fyne.Do(func() {
			ui.reconciliationStatementUI.Hide()
			dialog.ShowError(fmt.Errorf("error al reconciliar la cuenta: %v", err), ui.mainWindow)
		})
		return
	}

	ui.updateStatementCard(reconciliation)
	ui.reconciliationStatementUI.Show()
}

func (ui *UI) updateStatementCard(reconciliation *domain.Reconciliation) {
	// This function should update the reconciliation statement card with the reconciliation data.
	widgets := ui.reconciliationWidgets
	ui.reconciliationData = reconciliation

	// Update the labels with the reconciliation data
	widgets.endingDateLabel.SetText(fmt.Sprintf("Fecha de Cierre: %s",
		reconciliation.EndDate.Format("2006-01-02")))

	widgets.calculatedBalanceLabel.SetText(fmt.Sprintf("Saldo Calculado: $%s",
		reconciliation.CalculatedEndingBalance.StringFixed(2)))

	widgets.actualBalanceLabel.SetText(fmt.Sprintf("Saldo Real: $%s",
		reconciliation.EndingBalance.StringFixed(2)))

	widgets.differenceLabel.SetText(fmt.Sprintf("Diferencia: $%s",
		reconciliation.Difference.StringFixed(2)))

	// Update difference color
	bg := widgets.differenceContainer.Objects[0].(*canvas.Rectangle)
	if reconciliation.Difference.IsZero() {
		bg.FillColor = color.Transparent
		widgets.adjustmentButton.Disable()
	} else {
		bg.FillColor = color.NRGBA{R: 255, G: 0, B: 0, A: 60} // Light red
		widgets.adjustmentButton.Enable()
	}
	bg.Refresh()

	// Update the transaction list
	widgets.transactionList.Length = func() int {
		return len(ui.reconciliationData.Transactions)
	}
	widgets.transactionList.CreateItem = func() fyne.CanvasObject {
		// Create a template similar to your main transaction list item
		return container.NewGridWithColumns(4,
			widget.NewLabel("Date"),
			widget.NewLabel("Description"),
			widget.NewLabel("Type"),
			widget.NewLabel("Amount"),
		)
	}
	widgets.transactionList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		tx := ui.reconciliationData.Transactions[id]
		grid := item.(*fyne.Container)
		grid.Objects[0].(*widget.Label).SetText(tx.TransactionDate.Format("2006-01-02"))
		grid.Objects[1].(*widget.Label).SetText(tx.Description)
		grid.Objects[2].(*widget.Label).SetText(string(tx.Category.Type))
		grid.Objects[3].(*widget.Label).SetText(fmt.Sprintf("$%.2f", tx.Amount))
	}
	widgets.transactionList.Refresh()
}

func (ui *UI) makeStatementCard() fyne.CanvasObject {
	// Create the labels for the key figures
	endingDateLabel := widget.NewLabel("Fecha de Cierre: N/A")
	calculatedBalanceLabel := widget.NewLabel("Saldo Calculado: N/A")
	actualBalanceLabel := widget.NewLabel("Saldo Real: N/A")
	differenceLabel := widget.NewLabel("Diferencia: N/A")

	differenceContainer := container.NewStack(
		canvas.NewRectangle(color.Transparent),
		differenceLabel,
	)

	// Create the grid for the key figures
	keyFiguresGrid := container.NewGridWithColumns(2,
		endingDateLabel,
		calculatedBalanceLabel,
		actualBalanceLabel,
		differenceContainer,
	)

	// Create List for transactions
	transactionsList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)

	adjustmentButton := widget.NewButton("Crear Transacción de Ajuste", func() {
		dialogHandler := transaction.NewAdjustmentTransactionDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.TxService,
			ui.Services.CatService,
			func() {
				ui.reconciliationStatementUI.Hide()
			},
			ui.reconciliationData,
		)
		dialogHandler.Show()
	})
	adjustmentButton.Disable()

	finishButton := widget.NewButton("Finalizar Reconciliación", func() {
		ui.reconciliationStatementUI.Hide()
		ui.reconciliationData = nil
	})

	statementCard := widget.NewCard("Resultados de Reconciliación", "",
		container.NewBorder(keyFiguresGrid, container.NewHBox(adjustmentButton, finishButton), nil, nil, transactionsList),
	)

	ui.reconciliationWidgets = &reconciliationUIWidgets{
		endingDateLabel:        endingDateLabel,
		calculatedBalanceLabel: calculatedBalanceLabel,
		actualBalanceLabel:     actualBalanceLabel,
		differenceLabel:        differenceLabel,
		differenceContainer:    differenceContainer,
		transactionList:        transactionsList,
		adjustmentButton:       adjustmentButton,
	}

	return statementCard
}

func (ui *UI) loadAccountsForReconciliation(selector *widget.SelectEntry) {
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
		ui.accounts = accs
	}

	accountNames := make([]string, len(ui.accounts))
	for i, acc := range ui.accounts {
		accountNames[i] = acc.Name
	}

	selector.SetOptions(accountNames)
}

func formValidation(
	accountsSelector *widget.SelectEntry,
	endingDateEntry *widget.DateEntry,
	actualBalanceEntry *widget.Entry,
) {
	selectorValidatior := uivalidators.NewValidator()
	selectorValidatior.Required()
	accountsSelector.Validator = selectorValidatior.Validate

	dateValidator := uivalidators.NewValidator()
	dateValidator.IsDate()
	endingDateEntry.Validator = dateValidator.Validate

	balanceValidator := uivalidators.NewValidator()
	balanceValidator.IsFloat()
	actualBalanceEntry.Validator = balanceValidator.Validate
}
