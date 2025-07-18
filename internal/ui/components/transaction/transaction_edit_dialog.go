// Package transaction provides the dialog for editing transactions.
package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/category"
)

// EditTransactionDialog holds the state and logic for the 'Edit Transaction' dialog.
type EditTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	categoryService CategoryService
	callbackAction  func()
	txID            int

	// UI Components
	txNumber         *widget.Label
	descriptionEntry *widget.Entry
	amountEntry      *widget.Label
	dateEntry        *widget.DateEntry
	categoryLabel    *widget.Label
	categoryButton   *widget.Button
	searchDialog     *category.CategorySearchDialog
	attachmentLabel  *widget.Label
	searchFileBtn    *widget.Button

	// Data
	accountID          int
	selectedCategoryID int
	attachmentURI      fyne.URI
}

// NewEditTransactionDialog creates a new dialog handler for the edit action.
func NewEditTransactionDialog(win fyne.Window, l *log.Logger, txs TransactionService, cs CategoryService, callback func(), txID int, accountID int) *EditTransactionDialog {
	d := &EditTransactionDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		categoryService:  cs,
		callbackAction:   callback,
		txID:             txID,
		accountID:        accountID,
		txNumber:         widget.NewLabel(""),
		descriptionEntry: widget.NewMultiLineEntry(),
		amountEntry:      widget.NewLabel(""),
		dateEntry:        widget.NewDateEntry(),
		categoryLabel:    widget.NewLabel(""),
		attachmentLabel:  widget.NewLabel("Ninguno"),
	}
	d.categoryButton = widget.NewButtonWithIcon("", theme.SearchIcon(), d.openCategorySearch)
	d.searchDialog = category.NewCategorySearchDialog(win, l, cs, d.handleCategorySelect)
	d.searchFileBtn = widget.NewButtonWithIcon("", theme.FileIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, d.mainWin)
				return
			}
			if reader == nil {
				return
			}
			d.attachmentURI = reader.URI()
			d.attachmentLabel.SetText(reader.URI().Name())
		}, d.mainWin)
		fileDialog.Show()
	})
	return d
}

func (d *EditTransactionDialog) openCategorySearch() {
	d.searchDialog.Show()
}

func (d *EditTransactionDialog) handleCategorySelect(cat *domain.Category) {
	d.selectedCategoryID = cat.ID
	d.categoryLabel.SetText(cat.Name)
}

// Show begins the entire "edit" process.
func (d *EditTransactionDialog) Show() {
	onSuccess := func(tx *domain.Transaction) {
		fyne.Do(func() {
			d.showEditForm(tx)
		})
	}

	onFailure := func(err error) {
		d.logger.Println("Error getting transaction by ID:", err)
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error al cargar la transacción: %w", err), d.mainWin)
		})
	}

	d.fetchTransaction(onSuccess, onFailure)
}

func (d *EditTransactionDialog) fetchTransaction(onSuccess func(tx *domain.Transaction), onFailure func(err error)) {
	progress := dialog.NewCustomWithoutButtons("Cargando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		tx, err := d.txService.GetTransactionByID(ctx, d.txID)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
			})
			onFailure(err)
			return
		}

		fyne.Do(func() {
			progress.Hide()
		})
		onSuccess(tx)
	}()
}

func (d *EditTransactionDialog) showEditForm(tx *domain.Transaction) {
	d.txNumber.SetText(fmt.Sprintf("#%s", tx.TransactionNumber))
	d.descriptionEntry.SetText(tx.Description)
	d.amountEntry.SetText(fmt.Sprintf("%.2f", tx.Amount))
	d.dateEntry.SetText(tx.TransactionDate.Format("01/02/2006"))

	if tx.AttachmentPath != nil {
		uri, err := storage.ParseURI(*tx.AttachmentPath)
		if err == nil {
			d.attachmentURI = uri
			d.attachmentLabel.SetText(uri.Name())
		}
	}

	// This part requires a synchronous fetch or a callback to get category name
	// For simplicity, we'll assume category is loaded with transaction or fetched here
	// In a real app, you might pass the category name or fetch it.
	d.selectedCategoryID = tx.CategoryID
	// d.categoryLabel.SetText(tx.Category.Name) // Assuming tx.Category is populated

	categoryContainer := container.NewBorder(nil, nil, nil, d.categoryButton, d.categoryLabel)
	attachmentContainer := container.NewBorder(nil, nil, nil, d.searchFileBtn, d.attachmentLabel)
	txFormItems := NewTransactionForm(
		d.descriptionEntry,
		d.amountEntry,
		d.dateEntry,
		categoryContainer,
		attachmentContainer,
	)

	txNumberFormItem := widget.NewFormItem("Número de Transacción", d.txNumber)
	txFormItems = append([]*widget.FormItem{txNumberFormItem}, txFormItems...)

	formDialog := dialog.NewForm("Editar Transacción", "Guardar", "Cancelar",
		txFormItems,
		d.handleSubmit,
		d.mainWin,
	)
	formDialog.Resize(fyne.NewSize(560, 400))
	formDialog.Show()
}

func (d *EditTransactionDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	transactionDate, err := time.Parse("01/02/2006", d.dateEntry.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("formato de fecha inválido: %w", err), d.mainWin)
		return
	}

	if validator.IsDateAfter(transactionDate, time.Now()) {
		dialog.ShowError(errors.New("la fecha de la transacción no puede ser una fecha futura"), d.mainWin)
		return
	}

	progress := dialog.NewCustomWithoutButtons("Guardando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		amount, _ := strconv.ParseFloat(d.amountEntry.Text, 64)

		var attachmentPathPtr *string
		if d.attachmentURI != nil {
			uriString := d.attachmentURI.String()
			attachmentPathPtr = &uriString
		}

		updatedTx := &domain.Transaction{
			BaseEntity:      domain.BaseEntity{ID: d.txID},
			Description:     d.descriptionEntry.Text,
			Amount:          amount,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      d.selectedCategoryID,
			AttachmentPath:  attachmentPathPtr,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.txService.UpdateTransaction(ctx, updatedTx)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("error al editar la transacción: %w", err), d.mainWin)
			})
			d.logger.Printf("Error updating transaction %d: %v", d.txID, err)
			return
		}

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Exito!", "Cambios guardados correctamente", d.mainWin)
		})

		go d.callbackAction()
	}()
}
