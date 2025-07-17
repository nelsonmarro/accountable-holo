package transaction

import (
	"context"
	"errors"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// VoidTransactionDialog holds the dependencies for the delete confirmation dialog.
type VoidTransactionDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        TransactionService
	callbackAction func()
	txID           int
}

// NewVoidTransactionDialog creates a new dialog handler for the delete action.
func NewVoidTransactionDialog(win fyne.Window, l *log.Logger, service TransactionService, callback func(), txID int) *VoidTransactionDialog {
	return &VoidTransactionDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,
		txID:           txID,
	}
}

// Show displays the confirmation dialog to the user.
func (d *VoidTransactionDialog) Show() {
	confirmCallback := func(confirm bool) {
		if !confirm {
			return
		}
		d.executeVoid()
	}

	dialog.ShowConfirm(
		"Confirmar Anulación",
		"Está seguro que desea anular esta transacción? Esto creará una nueva transacción opuesta.",
		confirmCallback,
		d.mainWin,
	)
}

// executeVoid runs the actual deletion logic.
func (d *VoidTransactionDialog) executeVoid() {
	progress := dialog.NewCustomWithoutButtons("Anulando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.VoidTransaction(ctx, d.txID)
		if err != nil {
			d.logger.Printf("Error voiding transaction %d: %v", d.txID, err)
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(errors.New("error al anular la Transacción. Intente otra vez"), d.mainWin)
			})
			return
		}

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Success", "Transacción anuladada con éxito.", d.mainWin)
		})

		go d.callbackAction()
	}()
}
