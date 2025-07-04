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

// DeleteTransactionDialog holds the dependencies for the delete confirmation dialog.
type DeleteTransactionDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        TransactionService
	callbackAction func()
	txID           int
}

// NewDeleteTransactionDialog creates a new dialog handler for the delete action.
func NewDeleteTransactionDialog(win fyne.Window, l *log.Logger, service TransactionService, callback func(), txID int) *DeleteTransactionDialog {
	return &DeleteTransactionDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,
		txID:           txID,
	}
}

// Show displays the confirmation dialog to the user.
func (d *DeleteTransactionDialog) Show() {
	confirmCallback := func(confirm bool) {
		if !confirm {
			return
		}
		d.executeDelete()
	}

	dialog.ShowConfirm(
		"Confirm Void Transaction",
		"Are you sure you want to void this transaction? This will create a new, opposing transaction.",
		confirmCallback,
		d.mainWin,
	)
}

// executeDelete runs the actual deletion logic.
func (d *DeleteTransactionDialog) executeDelete() {
	progress := dialog.NewCustomWithoutButtons("Voiding...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		defer progress.Hide()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.VoidTransaction(ctx, d.txID)
		if err != nil {
			d.logger.Printf("Error voiding transaction %d: %v", d.txID, err)
			fyne.Do(func() {
				dialog.ShowError(errors.New("Error voiding transaction. Please try again."), d.mainWin)
			})
			return
		}

		fyne.Do(func() {
			dialog.ShowInformation("Success", "Transaction voided successfully.", d.mainWin)
		})

		go d.callbackAction()
	}()
}
