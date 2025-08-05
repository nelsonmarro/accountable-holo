package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// DeleteUserDialog holds the dependencies for the delete confirmation dialog.
type DeleteUserDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        UserService
	callbackAction func()
	userToDelete   *domain.User
	currentUser    *domain.User
}

// NewDeleteUserDialog creates a new dialog handler for the delete action.
func NewDeleteUserDialog(win fyne.Window, l *log.Logger, service UserService, callback func(), userToDelete *domain.User, currentUser *domain.User) *DeleteUserDialog {
	return &DeleteUserDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,
		userToDelete:   userToDelete,
		currentUser:    currentUser,
	}
}

// Show displays the confirmation dialog to the user.
func (d *DeleteUserDialog) Show() {
	confirmCallback := func(confirm bool) {
		if !confirm {
			return
		}
		d.executeDelete()
	}

	dialog.ShowConfirm(
		"Confirm Deletion",
		fmt.Sprintf("Are you sure you want to delete the user '%s'?", d.userToDelete.Username),
		confirmCallback,
		d.mainWin,
	)
}

// executeDelete runs the actual deletion logic.
func (d *DeleteUserDialog) executeDelete() {
	progress := dialog.NewCustomWithoutButtons("Deleting...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		defer progress.Hide()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.DeleteUser(ctx, d.userToDelete.ID, d.currentUser)
		if err != nil {
			d.logger.Printf("Error deleting user %d: %v", d.userToDelete.ID, err)
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("error deleting user: %w", err), d.mainWin)
			})
			return
		}

		fyne.Do(func() {
			dialog.ShowInformation("User Deleted", "User deleted successfully!", d.mainWin)
			go d.callbackAction()
		})
	}()
}
