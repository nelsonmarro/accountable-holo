package category

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

// DeleteCategoryDialog holds the dependencies for the delete confirmation dialog.
type DeleteCategoryDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        CategoryService
	callbackAction func()
	catID          int
}

// NewDeleteCategoryDialog creates a new dialog handler for the delete action.
func NewDeleteCategoryDialog(win fyne.Window, l *log.Logger, service CategoryService, callback func(), catID int) *DeleteCategoryDialog {
	return &DeleteCategoryDialog{
		mainWin: win, logger: l,
		service:        service,
		callbackAction: callback,
		catID:          catID,
	}
}

// Show displays the confirmation dialog to the user.
func (d *DeleteCategoryDialog) Show() {
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
		fmt.Sprintln("¿Está seguro de que desea eliminar la categoria? Esta acción no se puede deshacer."),
		confirmCallback,
		d.mainWin,
	)
}

// executeDelete runs the actual deletion logic.
func (d *DeleteCategoryDialog) executeDelete() {
	progress := dialog.NewCustomWithoutButtons("Eliminando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.DeleteCategory(ctx, d.catID)
		if err != nil {
			d.logger.Printf("Error deleting category %d: %v", d.catID, err)
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(errors.New("error al eliminar la categoria. Intente otra vez"), d.mainWin)
			})
			return
		}

		fyne.Do(func() {
			progress.Hide()
		})

		go d.callbackAction()
	}()
}
