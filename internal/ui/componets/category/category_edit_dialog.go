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
	"github.com/nelsonmarro/accountable-holo/internal/application/helpers"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// EditCategoryDialog holds the state and logic for the 'Edit Account' dialog.
type EditCategoryDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        CategoryService
	callbackAction func()
	catID          int

	// UI Components for the form
	nameEntry  *widget.Entry
	tipoSelect *widget.SelectEntry
}

// NewEditCategoryDialog creates a new dialog handler for the edit action.
func NewEditCategoryDialog(
	win fyne.Window,
	l *log.Logger,
	service CategoryService,
	callback func(),
	catID int,
) *EditCategoryDialog {
	return &EditCategoryDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,
		catID:          catID,
		// Initialize components
		nameEntry:  widget.NewEntry(),
		tipoSelect: widget.NewSelectEntry([]string{string(domain.Income), string(domain.Outcome)}),
	}
}

// Show begins the entire "edit" process.
func (d *EditCategoryDialog) Show() {
	// Define the function to run on successful data fetch.
	onSuccess := func(account *domain.Category) {
		fyne.Do(func() {
			d.showEditForm(account)
		})
	}

	// Define the function to run on failure.
	onFailure := func(err error) {
		d.logger.Println("Error getting category by ID:", err)
		fyne.Do(func() {
			errorDialog := dialog.NewError(fmt.Errorf("%s\n%s", "error al editar la categoria\n", err.Error()), d.mainWin)
			errorDialog.Show()
		})
	}

	// Start the asynchronous fetch process, passing our callbacks.
	d.fetchCategory(onSuccess, onFailure)
}

// Show begins the process by fetching the account data first.
func (d *EditCategoryDialog) fetchCategory(onSuccess func(acc *domain.Category), onFailure func(err error)) {
	progress := dialog.NewCustomWithoutButtons("Cargando Categoria...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		account, err := d.service.GetCategoryByID(ctx, d.catID)
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
func (d *EditCategoryDialog) showEditForm(acc *domain.Category) {
	// Populate the widgets with the fetched data
	d.nameEntry.SetText(acc.Name)
	d.tipoSelect.SetText(string(acc.Type))
	d.amountEntry.SetText(fmt.Sprintf("%.2f", acc.InitialBalance))
	d.amountEntry.Disable() // Initial balance should not be editable
	d.numberEntry.SetText(acc.Number)

	formDialog := dialog.NewForm("Editar Cuenta", "Guardar", "Cancelar",
		CategoryForm(
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
func (d *EditCategoryDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	progress := dialog.NewCustomWithoutButtons("Guardando Cambios...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		updatedCategory := &domain.Category{
			BaseEntity: domain.BaseEntity{
				ID: d.catID, // Use the original ID for updates
			},
			Name:   d.nameEntry.Text,
			Type:   helpers.GetCategoryTypeFromString(d.tipoSelect.Text),
			Number: d.numberEntry.Text,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.UpdateCategory(ctx, updatedCategory)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(errors.New("error al actualizar la cuenta. Intente otra vez"), d.mainWin)
			})
			d.logger.Printf("Error updating account %d: %v", d.catID, err)
			return
		}

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Éxito", "Los cambios han sido guardados.", d.mainWin)
		})

		go d.callbackAction()
	}()
}

