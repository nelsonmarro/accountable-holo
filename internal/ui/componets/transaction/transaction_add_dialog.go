package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
	"github.com/nelsonmarro/verith/internal/ui/componets/category"
	"github.com/nelsonmarro/verith/internal/ui/componets/taxpayer"
)

// AddTransactionDialog holds the state and logic for the 'Add Transaction' dialog.
type AddTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	recurService    RecurringTransactionService
	categoryService CategoryService
	taxService      TaxPayerService // Updated to local interface
	callbackAction  func()

	// UI Components
	dateEntry         *widget.DateEntry
	categoryLabel     *widget.Label
	searchCategoryBtn *widget.Button
	attachmentLabel   *widget.Label
	searchFileBtn     *widget.Button

	// Tax & Client UI
	subtotalLabel  *widget.Label
	taxAmountLabel *widget.Label
	totalLabel     *widget.Label

	// Client Selector
	taxPayerLabel     *widget.Label
	searchTaxPayerBtn *widget.Button

	// Maestro-Detalle
	itemsManager *ItemsListManager

	// Recurrence UI
	isRecurringCheck *widget.Check
	intervalSelect   *widget.Select

	// Data
	accountID        int
	selectedCategory *domain.Category
	selectedTaxPayer *domain.TaxPayer // Added
	attachmentPath   string
	currentUser      domain.User
	items            []domain.TransactionItem
}

// NewAddTransactionDialog creates a new dialog handler.
func NewAddTransactionDialog(
	win fyne.Window,
	l *log.Logger,
	txs TransactionService,
	rs RecurringTransactionService,
	cs CategoryService,
	ts TaxPayerService, // Added
	callback func(),
	accountID int,
	currentUser domain.User,
) *AddTransactionDialog {
	d := &AddTransactionDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		recurService:     rs,
		categoryService:  cs,
		taxService:       ts, // Added
		callbackAction:   callback,
		dateEntry:        widget.NewDateEntry(),
		accountID:        accountID,
		categoryLabel:    widget.NewLabel("Ninguna seleccionada"),
		attachmentLabel:  widget.NewLabel("Ninguno"),
		currentUser:      currentUser,
		isRecurringCheck: widget.NewCheck("", nil),
		intervalSelect:   widget.NewSelect([]string{"Mensual", "Semanal"}, nil),
		subtotalLabel:    widget.NewLabel("$0.00"),
		taxAmountLabel:   widget.NewLabel("$0.00"),
		totalLabel:       widget.NewLabelWithStyle("$0.00", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		taxPayerLabel:    widget.NewLabel("Consumidor Final"), // Default
		items:            make([]domain.TransactionItem, 0),
	}

	d.itemsManager = NewItemsListManager(win, d.handleItemsUpdate)

	d.dateEntry.SetText(time.Now().Format("01/02/2006"))
	d.intervalSelect.SetSelected("Mensual")
	d.intervalSelect.Hide()

	d.isRecurringCheck.OnChanged = func(checked bool) {
		if checked {
			d.intervalSelect.Show()
		} else {
			d.intervalSelect.Hide()
		}
	}

	d.searchCategoryBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchDialog := category.NewCategorySearchDialog(
			d.mainWin,
			d.logger,
			d.categoryService,
			func(cat *domain.Category) {
				d.selectedCategory = cat
				d.categoryLabel.SetText(cat.Name)

				// RESTRICCIÓN: Solo egresos pueden ser recurrentes
				if cat.Type == domain.Outcome {
					d.isRecurringCheck.Enable()
				} else {
					d.isRecurringCheck.SetChecked(false)
					d.isRecurringCheck.Disable()
					d.intervalSelect.Hide()
				}
			},
		)
		searchDialog.Show()
	})

	d.searchTaxPayerBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchDialog := taxpayer.NewSearchDialog(
			d.mainWin,
			d.logger,
			d.taxService,
			func(tp *domain.TaxPayer) {
				d.selectedTaxPayer = tp
				d.taxPayerLabel.SetText(tp.Name)
			},
		)
		searchDialog.Show()
	})

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

func (d *AddTransactionDialog) handleItemsUpdate(items []domain.TransactionItem) {
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

// Show creates and displays the Fyne form dialog.
func (d *AddTransactionDialog) Show() {
	categoryContainer := container.NewBorder(nil, nil, nil, d.searchCategoryBtn, d.categoryLabel)
	taxPayerContainer := container.NewBorder(nil, nil, nil, d.searchTaxPayerBtn, d.taxPayerLabel)
	attachmentContainer := container.NewBorder(nil, nil, nil, d.searchFileBtn, d.attachmentLabel)

	recurLabel := componets.NewHoverableLabel("¿Es Recurrente?", d.mainWin.Canvas())
	recurLabel.SetTooltip("Crea esta transacción automáticamente cada periodo (mes/semana).\nSe calculara a partir de la fecha de la transacción.")

	// Header Form
	headerForm := widget.NewForm(
		widget.NewFormItem("Cliente", taxPayerContainer),
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

	formDialog := dialog.NewCustomConfirm("Crear Transacción", "Guardar", "Cancelar",
		container.NewVScroll(container.NewPadded(content)),
		func(confirm bool) {
			if confirm {
				d.handleSubmit(true)
			}
		}, d.mainWin)

	formDialog.Resize(fyne.NewSize(850, 650)) // Larger for detail view
	formDialog.Show()
}

func (d *AddTransactionDialog) handleSubmit(_ bool) {
	if len(d.items) == 0 {
		dialog.ShowError(errors.New("debe agregar al menos un ítem"), d.mainWin)
		return
	}

	transactionDate, err := time.Parse("01/02/2006", d.dateEntry.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("formato de fecha inválido: %w", err), d.mainWin)
		return
	}

	if d.selectedCategory == nil {
		dialog.ShowError(fmt.Errorf("por favor, seleccione una categoría"), d.mainWin)
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Espere...", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		// Calculate final totals
		var sub15, sub0, taxTotal, total float64
		var description string

		for i, item := range d.items {
			if i > 0 {
				description += ", "
			}
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

		var taxPayerID *int
		if d.selectedTaxPayer != nil {
			taxPayerID = &d.selectedTaxPayer.ID
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		tx := &domain.Transaction{
			Description:     description,
			Amount:          total,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      d.selectedCategory.ID,
			AttachmentPath:  attachmentPathPtr,
			Subtotal15:      sub15,
			Subtotal0:       sub0,
			TaxAmount:       taxTotal,
			Items:           d.items,
			TaxPayerID:      taxPayerID, // Set ID
		}

		err := d.txService.CreateTransaction(ctx, tx, d.currentUser)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al crear la Transacción: %w", err), d.mainWin)
			})
			return
		}

		// 2. If Recurring (ONLY FOR OUTCOMES - UI already enforces this)
		if d.isRecurringCheck.Checked && d.selectedCategory.Type == domain.Outcome {
			intervalStr := d.intervalSelect.Selected
			var interval domain.RecurrenceInterval
			if intervalStr == "Semanal" {
				interval = domain.IntervalWeekly
			} else {
				interval = domain.IntervalMonthly
			}

			recurTx := &domain.RecurringTransaction{
				AccountID:   d.accountID,
				CategoryID:  d.selectedCategory.ID,
				Description: description,
				Amount:      total,
				StartDate:   transactionDate,
				Interval:    interval,
				IsActive:    true,
			}

			if err := d.recurService.Create(ctx, recurTx); err != nil {
				d.logger.Printf("Failed to create recurring transaction: %v", err)
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("transacción creada pero falló la recurrencia: %w", err), d.mainWin)
				})
			}
		}

		fyne.Do(func() {
			progressDialog.Hide()
			dialog.ShowInformation("Transacción Creada", "Transacción creada exitosamente!", d.mainWin)
		})

		go d.callbackAction()
	}()
}
