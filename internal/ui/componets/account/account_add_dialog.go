// Package account provides the UI components for managing accounts.
package account

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/helpers"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// AddAccountDialog holds the state and logic for the 'Add Account' dialog.
type AddAccountDialog struct {
	mainWin fyne.Window
	logger  *log.Logger
	service AccountService
	// callbackAction is called on successful creation to trigger a UI refresh.
	callbackAction func()

	// UI Components
	nameEntry   *widget.Entry
	tipoSelect  *widget.SelectEntry
	amountEntry *widget.Entry
	numberEntry *widget.Entry
}

// NewAddAccountDialog creates a new dialog handler.
func NewAddAccountDialog(win fyne.Window, l *log.Logger, service AccountService, callback func()) *AddAccountDialog {
	return &AddAccountDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,

		// Initialize components
		nameEntry:   widget.NewEntry(),
		tipoSelect:  widget.NewSelectEntry([]string{"Ahorros", "Corriente"}),
		amountEntry: widget.NewEntry(), // We can use your NumericEntry here later
		numberEntry: widget.NewEntry(),
	}
}

// Show creates and displays the Fyne form dialog.
func (d *AddAccountDialog) Show() {
	formDialog := dialog.NewForm("Crear Cuenta", "Guardar", "Cancelar",
		AccountForm(
			d.nameEntry,
			d.tipoSelect,
			d.amountEntry,
			d.numberEntry,
		),
		d.handleSubmit, // Pass the method as the callback
		d.mainWin,
	)

	formDialog.Resize(fyne.NewSize(480, 300))
	formDialog.Show()
}

func (d *AddAccountDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	name := d.nameEntry.Text
	tipo := d.tipoSelect.Text
	amount, _ := strconv.ParseFloat(d.amountEntry.Text, 64)
	number := d.numberEntry.Text

	progressDialog := dialog.NewCustomWithoutButtons("Espere", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		acc := &domain.Account{
			Name:           name,
			Type:           helpers.GetAccountTypeFromString(tipo),
			Number:         number,
			InitialBalance: amount,
		}

		cxt, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.CreateNewAccount(cxt, acc)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				errorDialog := dialog.NewError(fmt.Errorf("%s\n%s", "error al crear la cuenta\n", err.Error()), d.mainWin)
				errorDialog.Show()
			})
			d.logger.Println("Error creating account:", err)
			return
		}

		fyne.Do(func() {
			progressDialog.Hide()
			infoDialog := dialog.NewInformation("Cuenta Creada",
				fmt.Sprintf("Cuenta: %s - %s\nTipo: %s\nSaldo inicial: %.2f\nCreada!", name, number, tipo, amount), d.mainWin)
			infoDialog.Show()
		})

		go d.callbackAction()
	}()
}
