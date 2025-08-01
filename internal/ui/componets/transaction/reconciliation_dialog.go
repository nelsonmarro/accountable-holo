package transaction

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

type reconciliationUIWidgets struct {
	endingDateLabel        *widget.Label
	calculatedBalanceLabel *widget.Label
	actualBalanceLabel     *widget.Label
	differenceLabel        *widget.Label
	differenceContainer    *fyne.Container
	transactionList        *widget.List
	adjustmentButton       *widget.Button
}

type ReconciliationDialog struct {
	dialog                dialog.Dialog
	TxService             TransactionService
	CatService            CategoryService
	logger                *log.Logger
	mainWindow            fyne.Window
	statementUI           fyne.CanvasObject
	data                  *domain.Reconciliation
	widgets               *reconciliationUIWidgets
	accounts              []domain.Account
	onAdjustmentTxCreated func()
}

func NewReconciliationDialog(
	mainWindow fyne.Window,
	logger *log.Logger,
	txService TransactionService,
	catService CategoryService,
	accounts []domain.Account,
	onAdjustmentTxCreated func(),
) *ReconciliationDialog {
	d := &ReconciliationDialog{
		TxService:             txService,
		CatService:            catService,
		logger:                logger,
		mainWindow:            mainWindow,
		accounts:              accounts,
		onAdjustmentTxCreated: onAdjustmentTxCreated,
	}
	d.statementUI = d.makeStatementCard()
	d.statementUI.Hide()

	formCard := d.makeFormCard()
	dialogContent := container.NewBorder(formCard, nil, nil, nil, d.statementUI)
	d.dialog = dialog.NewCustom("Reconciliación de Cuenta", "Cerrar", dialogContent, mainWindow)
	d.dialog.Resize(fyne.NewSize(800, 600))

	return d
}

func (d *ReconciliationDialog) Show() {
	d.dialog.Show()
}

func (d *ReconciliationDialog) makeFormCard() fyne.CanvasObject {
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
		for _, acc := range d.accounts {
			if acc.Name == selectedAccountName {
				selectedAccountID = acc.ID
				break
			}
		}

		if selectedAccountID == 0 {
			dialog.ShowError(fmt.Errorf("la cuenta seleccionada no es válida"), d.mainWindow)
			return
		}

		endingDate := endingDateEntry.Date

		actualBalance, err := decimal.NewFromString(actualBalanceEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("saldo final real no es un número válido: %v", err), d.mainWindow)
			return
		}

		go d.initiateReconciliation(selectedAccountID, endingDate, actualBalance)
	}

	// Create the card itself
	formCard := widget.NewCard(
		"Reconciliación de Cuenta",
		"",
		reconciliationForm,
	)

	// Don't forget to load the accounts for the selector, similar to how you do it in the
	d.loadAccountsForReconciliation(accountsSelector)

	return formCard
}

func (d *ReconciliationDialog) initiateReconciliation(accountID int, endingDate *time.Time, actualBalance decimal.Decimal) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reconciliation, err := d.TxService.ReconcileAccount(ctx, accountID, *endingDate, actualBalance)
	if err != nil {
		fyne.Do(func() {
			d.statementUI.Hide()
			dialog.ShowError(fmt.Errorf("error al reconciliar la cuenta: %v", err), d.mainWindow)
		})
		return
	}

	d.data = reconciliation
	d.updateStatementCard()
	d.statementUI.Show()
}

// updateStatementCard updates the reconciliation statement card with the latest data.
func (d *ReconciliationDialog) updateStatementCard() {
	d.widgets.endingDateLabel.SetText(fmt.Sprintf("Fecha de Cierre: %s",
		d.data.EndDate.Format("2006-01-02")))

	d.widgets.calculatedBalanceLabel.SetText(fmt.Sprintf("Saldo Calculado: $%s",
		d.data.CalculatedEndingBalance.StringFixed(2)))

	d.widgets.actualBalanceLabel.SetText(fmt.Sprintf("Saldo Real: $%s",
		d.data.EndingBalance.StringFixed(2)))

	d.widgets.differenceLabel.SetText(fmt.Sprintf("Diferencia: $%s",
		d.data.Difference.StringFixed(2)))

	// Update difference color
	bg := d.widgets.differenceContainer.Objects[0].(*canvas.Rectangle)
	if d.data.Difference.IsZero() {
		bg.FillColor = color.Transparent
		d.widgets.adjustmentButton.Disable()
	} else {
		bg.FillColor = color.NRGBA{R: 255, G: 0, B: 0, A: 60} // Light red
		d.widgets.adjustmentButton.Enable()
	}
	bg.Refresh()

	// Update the transaction list
	d.widgets.transactionList.Length = func() int {
		return len(d.data.Transactions)
	}
	d.widgets.transactionList.CreateItem = func() fyne.CanvasObject {
		// Create a template similar to your main transaction list item
		return container.NewGridWithColumns(4,
			widget.NewLabel("Date"),
			widget.NewLabel("Description"),
			widget.NewLabel("Type"),
			widget.NewLabel("Amount"),
		)
	}
	d.widgets.transactionList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		tx := d.data.Transactions[id]
		grid := item.(*fyne.Container)
		grid.Objects[0].(*widget.Label).SetText(tx.TransactionDate.Format("2006-01-02"))
		grid.Objects[1].(*widget.Label).SetText(tx.Description)
		grid.Objects[2].(*widget.Label).SetText(string(tx.Category.Type))
		grid.Objects[3].(*widget.Label).SetText(fmt.Sprintf("$%.2f", tx.Amount))
	}
	d.widgets.transactionList.Refresh()
}

func (d *ReconciliationDialog) makeStatementCard() fyne.CanvasObject {
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
		dialogHandler := NewAdjustmentTransactionDialog(
			d.mainWindow,
			d.logger,
			d.TxService,
			d.CatService,
			d.data,
			d.onAdjustmentTxCreated,
		)
		dialogHandler.Show()
	})
	adjustmentButton.Disable()

	finishButton := widget.NewButton("Finalizar Reconciliación", func() {
		d.statementUI.Hide()
		d.data = nil
		d.dialog.Hide()
	})

	statementCard := widget.NewCard("Resultados de Reconciliación", "",
		container.NewBorder(keyFiguresGrid, container.NewHBox(adjustmentButton, finishButton), nil, nil, transactionsList),
	)

	d.widgets = &reconciliationUIWidgets{
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

func (d *ReconciliationDialog) loadAccountsForReconciliation(selector *widget.SelectEntry) {
	if d.accounts == nil {
		dialog.ShowError(fmt.Errorf("las cuentas no se pudieron cargar"), d.mainWindow)
		return
	}

	accountNames := make([]string, len(d.accounts))
	for i, acc := range d.accounts {
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
