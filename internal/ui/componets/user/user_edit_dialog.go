// Package user provides the UI components related to user management.
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

// EditUserDialog holds the state and logic for the 'Edit User' dialog.
type EditUserDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	userService    UserService
	callbackAction func()
	currentUser    *domain.User
	userToEdit     *domain.User

	// UI Components
	usernameEntry *widget.Entry
	passwordEntry *widget.Entry
	roleSelect    *widget.SelectEntry
}

// NewEditUserDialog creates a new dialog handler.
func NewEditUserDialog(
	win fyne.Window,
	l *log.Logger,
	us UserService,
	callback func(),
	currentUser *domain.User,
	userToEdit *domain.User,
) *EditUserDialog {
	d := &EditUserDialog{
		mainWin:        win,
		logger:         l,
		userService:    us,
		callbackAction: callback,
		currentUser:    currentUser,
		userToEdit:     userToEdit,
		usernameEntry:  widget.NewEntry(),
		passwordEntry:  &widget.Entry{Password: true},
		roleSelect:     widget.NewSelectEntry([]string{string(domain.AdminRole), string(domain.CustomerRole)}),
	}
	d.usernameEntry.SetText(userToEdit.Username)
	d.roleSelect.SetText(string(userToEdit.Role))
	return d
}

// Show creates and displays the Fyne form dialog.
func (d *EditUserDialog) Show() {
	formDialog := dialog.NewForm("Edit User", "Save", "Cancel",
		UserForm(
			d.usernameEntry,
			d.passwordEntry,
			d.roleSelect,
		),
		d.handleSubmit,
		d.mainWin,
	)

	formDialog.Resize(fyne.NewSize(400, 250))
	formDialog.Show()
}

func (d *EditUserDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Please wait...", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		defer progressDialog.Hide()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.userService.UpdateUser(ctx, d.userToEdit.ID, d.usernameEntry.Text, d.passwordEntry.Text, domain.UserRole(d.roleSelect.Text), d.currentUser)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("error updating user: %w", err), d.mainWin)
			})
			d.logger.Println("Error updating user:", err)
			return
		}

		fyne.Do(func() {
			dialog.ShowInformation("User Updated", "User updated successfully!", d.mainWin)
			go d.callbackAction()
		})
	}()
}
