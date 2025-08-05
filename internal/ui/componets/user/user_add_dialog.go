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

// AddUserDialog holds the state and logic for the 'Add User' dialog.
type AddUserDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	userService    UserService
	callbackAction func()
	currentUser    *domain.User

	// UI Components
	usernameEntry *widget.Entry
	passwordEntry *widget.Entry
	roleSelect    *widget.SelectEntry
}

// NewAddUserDialog creates a new dialog handler.
func NewAddUserDialog(
	win fyne.Window,
	l *log.Logger,
	us UserService,
	callback func(),
	currentUser *domain.User,
) *AddUserDialog {
	d := &AddUserDialog{
		mainWin:        win,
		logger:         l,
		userService:    us,
		callbackAction: callback,
		currentUser:    currentUser,
		usernameEntry:  widget.NewEntry(),
		passwordEntry:  &widget.Entry{Password: true},
		roleSelect:     widget.NewSelectEntry([]string{string(domain.AdminRole), string(domain.CustomerRole)}),
	}
	d.roleSelect.SetText(string(domain.CustomerRole)) // Default to Customer
	return d
}

// Show creates and displays the Fyne form dialog.
func (d *AddUserDialog) Show() {
	formDialog := dialog.NewForm("Create User", "Save", "Cancel",
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

func (d *AddUserDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Please wait...", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		defer progressDialog.Hide()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.userService.CreateUser(ctx, d.usernameEntry.Text, d.passwordEntry.Text, domain.UserRole(d.roleSelect.Text), d.currentUser)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("error creating user: %w", err), d.mainWin)
			})
			d.logger.Println("Error creating user:", err)
			return
		}

		fyne.Do(func() {
			dialog.ShowInformation("User Created", "User created successfully!", d.mainWin)
			go d.callbackAction()
		})
	}()
}
