package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
)

// AddUserDialog holds the state and logic for the 'Add User' dialog.
type AddUserDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	userService    UserService
	callbackAction func()
	currentUser    *domain.User

	// UI Components
	usernameEntry  *widget.Entry
	passwordEntry  *widget.Entry
	firstNameEntry *widget.Entry
	lastNameEntry  *widget.Entry
	roleSelect     *widget.SelectEntry
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
		firstNameEntry: widget.NewEntry(),
		lastNameEntry:  widget.NewEntry(),
		roleSelect:     widget.NewSelectEntry([]string{string(domain.RoleAdmin), string(domain.RoleSupervisor), string(domain.RoleCashier)}),
	}
	d.roleSelect.SetText(string(domain.RoleCashier)) // Default to Cashier
	return d
}

// Show creates and displays the Fyne form dialog.
func (d *AddUserDialog) Show() {
	roleDesc := widget.NewLabel("")
	roleDesc.Wrapping = fyne.TextWrapWord
	
	updateDesc := func(role string) {
		switch domain.UserRole(role) {
		case domain.RoleAdmin:
			roleDesc.SetText("Control total: Configuración SRI, Usuarios, Reportes y Finanzas.")
		case domain.RoleSupervisor:
			roleDesc.SetText("Gestión Táctica: Reportes, Reconciliación y Anulaciones. Sin acceso a configuración.")
		case domain.RoleCashier:
			roleDesc.SetText("Operativo: Ventas y Clientes. Sin acceso a reportes financieros.")
		default:
			roleDesc.SetText("Seleccione un rol para ver sus permisos.")
		}
	}

	d.roleSelect.OnChanged = func(s string) {
		updateDesc(s)
	}
	// Initial state
	updateDesc(d.roleSelect.Text)

	formDialog := dialog.NewForm("Create User", "Save", "Cancel",
		UserForm(
			d.usernameEntry,
			d.passwordEntry,
			d.firstNameEntry,
			d.lastNameEntry,
			d.roleSelect,
			roleDesc,
		),
		d.handleSubmit,
		d.mainWin,
	)

	formDialog.Resize(fyne.NewSize(450, 400)) // Un poco más alto para la descripción
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

		err := d.userService.CreateUser(ctx, d.usernameEntry.Text, d.passwordEntry.Text, d.firstNameEntry.Text, d.lastNameEntry.Text, domain.UserRole(d.roleSelect.Text), d.currentUser)
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
