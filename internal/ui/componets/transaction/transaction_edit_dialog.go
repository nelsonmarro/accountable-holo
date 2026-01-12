// Package transaction provides the dialog for editing transactions.
package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/category"
)

// EditTransactionDialog holds the state and logic for the 'Edit Transaction' dialog.
type EditTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	recurService    RecurringTransactionService
	categoryService CategoryService
	callbackAction  func()
	txID            int

	// UI Components
	txNumber         *widget.Label
	descriptionEntry *widget.Entry // Derived/Generated description
	dateEntry        *widget.DateEntry
	categoryLabel    *widget.Label
	categoryButton   *widget.Button
	searchDialog     *category.CategorySearchDialog
	attachmentLabel  *widget.Label
	searchFileBtn    *widget.Button
	
	// Tax & Client UI
	subtotalLabel  *widget.Label
	taxAmountLabel *widget.Label
	totalLabel     *widget.Label
	taxPayerEntry  *widget.Entry 

	// Maestro-Detalle
	itemsManager *ItemsListManager

	// Recurrence UI
	isRecurringCheck *widget.Check
	intervalSelect   *widget.Select

	// Data
	accountID          int
	categories         []domain.Category
	selectedCategoryID int
	attachmentPath     string
	currentUser        domain.User
	items              []domain.TransactionItem
}

// NewEditTransactionDialog creates a new dialog handler for the edit action.
func NewEditTransactionDialog(
	win fyne.Window,
	l *log.Logger,
	txs TransactionService,
	rs RecurringTransactionService,
	cs CategoryService,
	callback func(),
	txID int,
	accountID int,
	currentUser domain.User,
) *EditTransactionDialog {
	d := &EditTransactionDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		recurService:     rs,
		categoryService:  cs,
		callbackAction:   callback,
		txID:             txID,
		accountID:        accountID,
		currentUser:      currentUser,
		txNumber:         widget.NewLabel(""),
		descriptionEntry: widget.NewMultiLineEntry(),
		dateEntry:        widget.NewDateEntry(),
		categoryLabel:    widget.NewLabel(""),
		attachmentLabel:  widget.NewLabel("Ninguno"),
		isRecurringCheck: widget.NewCheck("", nil),
		intervalSelect:   widget.NewSelect([]string{"Mensual", "Semanal"}, nil),
		subtotalLabel:    widget.NewLabel("$0.00"),
		taxAmountLabel:   widget.NewLabel("$0.00"),
		totalLabel:       widget.NewLabelWithStyle("$0.00", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		taxPayerEntry:    widget.NewEntry(),
		items:            make([]domain.TransactionItem, 0),
	}
	d.intervalSelect.SetSelected("Mensual")
	d.intervalSelect.Hide()
	d.taxPayerEntry.SetPlaceHolder("RUC / Cédula Cliente (Opcional)")

	d.itemsManager = NewItemsListManager(win, d.handleItemsUpdate)

	d.isRecurringCheck.OnChanged = func(checked bool) {
		if checked {
			d.intervalSelect.Show()
		} else {
			d.intervalSelect.Hide()
		}
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
			d.attachmentPath = reader.URI().Path()
			d.attachmentLabel.SetText(reader.URI().Name())
		}, d.mainWin)
		fileDialog.Show()
	})
	return d
}

func (d *EditTransactionDialog) handleItemsUpdate(items []domain.TransactionItem) {
	d.items = items
	var subtotal, tax float64
	
	for _, item := range items {
		subtotal += item.Subtotal
		if item.TaxRate == 4 { // IVA 15%
			tax += item.Subtotal * 0.15
		}
	}
	
	d.subtotalLabel.SetText(fmt.Sprintf("$%.2f", subtotal))
	d.taxAmountLabel.SetText(fmt.Sprintf("$%.2f", tax))
	d.totalLabel.SetText(fmt.Sprintf("$%.2f", subtotal+tax))
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
	onSuccess := func(tx *domain.Transaction, items []domain.TransactionItem) {
		fyne.Do(func() {
			d.showEditForm(tx, items)
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

func (d *EditTransactionDialog) fetchTransaction(onSuccess func(tx *domain.Transaction, items []domain.TransactionItem), onFailure func(err error)) {
	progress := dialog.NewCustomWithoutButtons("Cargando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		tx, err := d.txService.GetTransactionByID(ctx, d.txID)
		if err != nil {
			fyne.Do(func() { progress.Hide() })
			onFailure(err)
			return
		}

		items, err := d.txService.GetItemsByTransactionID(ctx, d.txID)
		if err != nil {
			fyne.Do(func() { progress.Hide() })
			onFailure(err)
			return
		}

		categories, err := d.categoryService.GetAllCategories(ctx)
		if err != nil {
			fyne.Do(func() { progress.Hide() })
			onFailure(err)
			return
		}
		d.categories = categories

		fyne.Do(func() { progress.Hide() })
		onSuccess(tx, items)
	}()
}

func (d *EditTransactionDialog) showEditForm(tx *domain.Transaction, items []domain.TransactionItem) {
	d.txNumber.SetText(fmt.Sprintf("#%s", tx.TransactionNumber))
	d.descriptionEntry.SetText(tx.Description)
	d.dateEntry.SetText(tx.TransactionDate.Format("01/02/2006"))

	if tx.AttachmentPath != nil {
		d.attachmentPath = *tx.AttachmentPath
		d.attachmentLabel.SetText(filepath.Base(*tx.AttachmentPath))
	}

	if tx.TaxPayerID != nil {
		d.taxPayerEntry.SetText(strconv.Itoa(*tx.TaxPayerID))
	}

	for _, cat := range d.categories {
		if cat.ID == tx.CategoryID {
			d.selectedCategoryID = cat.ID
			d.categoryLabel.SetText(cat.Name)
			break
		}
	}
	
	d.itemsManager.SetItems(items)

	categoryContainer := container.NewBorder(nil, nil, nil, d.categoryButton, d.categoryLabel)
	attachmentContainer := container.NewBorder(nil, nil, nil, d.searchFileBtn, d.attachmentLabel)

	recurLabel := componets.NewHoverableLabel("¿Es Recurrente?", d.mainWin.Canvas())
	recurLabel.SetTooltip("Convierte esta transacción en recurrente (se creará una nueva regla)")

	// Header Form
	headerForm := widget.NewForm(
		widget.NewFormItem("Número", d.txNumber),
		widget.NewFormItem("Cliente", d.taxPayerEntry),
		widget.NewFormItem("Fecha", d.dateEntry),
		widget.NewFormItem("Categoría", categoryContainer),
		widget.NewFormItem("Adjunto", attachmentContainer),
		widget.NewFormItem("", container.NewBorder(nil, nil, recurLabel, nil, d.isRecurringCheck)),
		widget.NewFormItem("Frecuencia", d.intervalSelect),
	)

	// Summary / Totals
	summary := widget.NewForm(
		widget.NewFormItem("Subtotal", d.subtotalLabel),
		widget.NewFormItem("IVA", d.taxAmountLabel),
		widget.NewFormItem("TOTAL", d.totalLabel),
	)

	// Main Layout
	content := container.NewVBox(
		headerForm,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Detalle de Ítems", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		d.itemsManager.GetContent(),
		widget.NewSeparator(),
		summary,
	)

	formDialog := dialog.NewCustomConfirm("Editar Transacción", "Guardar", "Cancelar",
		container.NewVScroll(container.NewPadded(content)),
		func(confirm bool) {
			if confirm {
				d.handleSubmit(true)
			}
		}, d.mainWin,
	)
	formDialog.Resize(fyne.NewSize(650, 600))
	formDialog.Show()
}

func (d *EditTransactionDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}
	
	if len(d.items) == 0 {
		dialog.ShowError(errors.New("debe tener al menos un ítem"), d.mainWin)
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
		// Calculate final totals
		var sub15, sub0, taxTotal, total float64
		var description string
		
		for i, item := range d.items {
			if i > 0 { description += ", " }
			description += item.Description
			
			if item.TaxRate == 4 {
				sub15 += item.Subtotal
				taxTotal += item.Subtotal * 0.15
			} else {
				sub0 += item.Subtotal
			}
		}
		total = sub15 + sub0 + taxTotal

		var attachmentPathPtr *string
		if d.attachmentPath != "" {
			attachmentPathPtr = &d.attachmentPath
		}

		updatedTx := &domain.Transaction{
			BaseEntity:      domain.BaseEntity{ID: d.txID},
			Description:     description, // Auto-generated from items
			Amount:          total,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      d.selectedCategoryID,
			AttachmentPath:  attachmentPathPtr,
			Subtotal15:      sub15,
			Subtotal0:       sub0,
			TaxAmount:       taxTotal,
			Items:           d.items, // Needs backend support to update items
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// NOTE: TransactionService.UpdateTransaction needs to be updated to handle Items update (Delete old + Insert new)
		err := d.txService.UpdateTransaction(ctx, updatedTx, d.currentUser)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("error al editar la transacción: %w", err), d.mainWin)
			})
			d.logger.Printf("Error updating transaction %d: %v", d.txID, err)
			return
		}

		// Handle Recurrence Creation
		if d.isRecurringCheck.Checked {
			interval := domain.IntervalMonthly
			if d.intervalSelect.Selected == "Semanal" {
				interval = domain.IntervalWeekly
			}

			nextRun := transactionDate
			if interval == domain.IntervalMonthly {
				nextRun = nextRun.AddDate(0, 1, 0)
			} else {
				nextRun = nextRun.AddDate(0, 0, 7)
			}

			rt := &domain.RecurringTransaction{
				Description: description,
				Amount:      total,
				AccountID:   d.accountID,
				CategoryID:  d.selectedCategoryID,
				Interval:    interval,
				StartDate:   transactionDate,
				NextRunDate: nextRun,
				IsActive:    true,
			}

			if err := d.recurService.Create(ctx, rt); err != nil {
				d.logger.Printf("Failed to create recurrence from edit: %v", err)
				fyne.Do(func() {
					dialog.ShowInformation("Advertencia", "Transacción actualizada, pero falló la creación de la recurrencia.", d.mainWin)
				})
			}
		}

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Exito!", "Cambios guardados correctamente", d.mainWin)
		})

		go d.callbackAction()
	}()
}