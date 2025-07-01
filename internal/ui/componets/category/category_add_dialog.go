// Package category has the UI components for cateogry
package category

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

// AddCategoryDialog holds the state and logic for the 'Add Account' dialog.
type AddCategoryDialog struct {
	mainWin fyne.Window
	logger  *log.Logger
	service CategoryService
	// callbackAction is called on successful creation to trigger a UI refresh.
	callbackAction func()

	// UI Components
	nameEntry   *widget.Entry
	tipoSelect  *widget.SelectEntry
	amountEntry *widget.Entry
	numberEntry *widget.Entry
}

// NewAddCategoryDialog creates a new dialog handler.
func NewAddCategoryDialog(win fyne.Window, l *log.Logger, service CategoryService, callback func()) *AddCategoryDialog {
	return &AddCategoryDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		callbackAction: callback,

		// Initialize components
		nameEntry:  widget.NewEntry(),
		tipoSelect: widget.NewSelectEntry([]string{string(domain.Income), string(domain.Outcome)}),
	}
}

// Show creates and displays the Fyne form dialog.
func (d *AddCategoryDialog) Show() {
	formDialog := dialog.NewForm("Crear Categoria", "Guardar", "Cancelar",
		CategoryForm(
			d.nameEntry,
			d.tipoSelect,
		),
		d.handleSubmit, // Pass the method as the callback
		d.mainWin,
	)

	formDialog.Resize(fyne.NewSize(380, 200))
	formDialog.Show()
}

func (d *AddCategoryDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	name := d.nameEntry.Text
	tipo := d.tipoSelect.Text

	progressDialog := dialog.NewCustomWithoutButtons("Espere", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		cat := &domain.Category{
			Name: name,
			Type: helpers.GetCategoryTypeFromString(tipo),
		}

		cxt, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.service.CreateCategory(cxt, cat)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				errorDialog := dialog.NewError(fmt.Errorf("%s\n%s", "error al crear la categoria\n", err.Error()), d.mainWin)
				errorDialog.Show()
			})
			d.logger.Println("Error creating category:", err)
			return
		}

		fyne.Do(func() {
			progressDialog.Hide()
			infoDialog := dialog.NewInformation("Categoria Creada",
				fmt.Sprintf("Categoria: %s \nTipo: %s\nCreada!", name, tipo), d.mainWin)
			infoDialog.Show()
		})

		go d.callbackAction()
	}()
}
