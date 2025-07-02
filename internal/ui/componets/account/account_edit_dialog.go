package account

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/helpers"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// EditAccountDialog holds the state and logic for the 'Edit Account' dialog.
type EditAccountDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        AccountService
	callbackAction func()
	accountID      int

	// UI Components for the form
	nameEntry   *widget.Entry
	tipoSelect  *widget.SelectEntry
	amountEntry *widget.Entry
	numberEntry *widget.Entry
}

// NewEditAccountDialog creates a new dialog handler for the edit action.
func NewEditAccountDialog(win fyne.Window, l *log.Logger, service AccountService, callback func(), accID int) *EditAccountDialog {
	return &EditAccountDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,
		accountID:      accID,
		// Initialize components
		nameEntry:   widget.NewEntry(),
		tipoSelect:  widget.NewSelectEntry([]string{string(domain.SavingAccount), string(domain.OrdinaryAccount)}),
		amountEntry: widget.NewEntry(),
		numberEntry: widget.NewEntry(),
	}
}

// Show begins the entire "edit" process.
func (d *EditAccountDialog) Show() {
	// Define the function to run on successful data fetch.
	onSuccess := func(account *domain.Account) {
		fyne.Do(func() {
			d.showEditForm(account)
		})
	}

	// Define the function to run on failure.
	onFailure := func(err error) {
		d.logger.Println("Error getting account by ID:", err)
		fyne.Do(func() {
			errorDialog := dialog.NewError(fmt.Errorf("%s\n%s", "error al editar la cuenta", err.Error()), d.mainWin)
			errorDialog.Show()
		})
	}

	// Start the asynchronous fetch process, passing our callbacks.
	d.fetchAccount(onSuccess, onFailure)
}

// Show begins the process by fetching the account data first.
func (d *EditAccountDialog) fetchAccount(onSuccess func(acc *domain.Account), onFailure func(err error)) {
	progress := dialog.NewCustomWithoutButtons("Cargando Cuenta...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		account, err := d.service.GetAccountByID(ctx, d.accountID)
		if err != nil {
			// If there's an error, call the failure callback.
			fyne.Do(func() {
				progress.Hide()
			})
			onFailure(err)
			return
		}

		// If successful, call the success callback with the fetched data.
		fyne.Do(func() {
			progress.Hide()
		})
		onSuccess(account)
	}()
}

// showEditForm displays the actual form, pre-populated with account data.
func (d *EditAccountDialog) showEditForm(acc *domain.Account) {
	// Populate the widgets with the fetched data
	d.nameEntry.SetText(acc.Name)
	d.tipoSelect.SetText(string(acc.Type))
	d.amountEntry.SetText(fmt.Sprintf("%.2f", acc.InitialBalance))
	d.amountEntry.Disable() // Initial balance should not be editable
	d.numberEntry.SetText(acc.Number)

	formDialog := dialog.NewForm("Editar Cuenta", "Guardar", "Cancelar",
		AccountForm(
			d.nameEntry,
			d.tipoSelect,
			d.amountEntry,
			d.numberEntry,
		),
		d.handleSubmit, // The submit callback
		d.mainWin,
	)
	formDialog.Resize(fyne.NewSize(480, 300))
	formDialog.Show()
}

// handleSubmit contains the logic for the UPDATE operation.
func (d *EditAccountDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	progress := dialog.NewCustomWithoutButtons("Guardando Cambios...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		updatedAccount := &domain.Account{
			BaseEntity: domain.BaseEntity{
				ID: d.accountID, // Use the original ID for updates
			},
			Name:   d.nameEntry.Text,
			Type:   helpers.GetAccountTypeFromString(d.tipoSelect.Text),
			Number: d.numberEntry.Text,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.UpdateAccount(ctx, updatedAccount)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(err, d.mainWin)
			})
			d.logger.Printf("Error updating account %d: %v", d.accountID, err)
			return
		}

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Éxito", "Los cambios han sido guardados.", d.mainWin)
		})

		go d.callbackAction()
	}()
}
