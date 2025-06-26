package account

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// DeleteAccountDialog holds the dependencies for the delete confirmation dialog.
type DeleteAccountDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        AccountService
	callbackAction func()
	accountID      int
}

// NewDeleteAccountDialog creates a new dialog handler for the delete action.
func NewDeleteAccountDialog(win fyne.Window, l *log.Logger, service AccountService, callback func(), accID int) *DeleteAccountDialog {
	return &DeleteAccountDialog{
		mainWin: win, logger: l,
		service:        service,
		callbackAction: callback,
		accountID:      accID,
	}
}

// Show displays the confirmation dialog to the user.
func (d *DeleteAccountDialog) Show() {
	confirmCallback := func(confirm bool) {
		if !confirm {
			// User clicked "No", so we do nothing.
			return
		}
		// User clicked "Yes", so we proceed with deletion in a goroutine.
		d.executeDelete()
	}

	dialog.ShowConfirm(
		"Confirmar Eliminación",
		fmt.Sprintln("¿Está seguro de que desea eliminar la cuenta? Esta acción no se puede deshacer."),
		confirmCallback,
		d.mainWin,
	)
}

// executeDelete runs the actual deletion logic.
func (d *DeleteAccountDialog) executeDelete() {
	progress := dialog.NewCustomWithoutButtons("Eliminando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.DeleteAccount(ctx, d.accountID)
		if err != nil {
			d.logger.Printf("Error deleting account %d: %v", d.accountID, err)
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(errors.New("error al eliminar la cuenta. Intente otra vez"), d.mainWin)
			})
			return
		}

		fyne.Do(func() {
			progress.Hide()
		})

		go d.callbackAction()
	}()
}
