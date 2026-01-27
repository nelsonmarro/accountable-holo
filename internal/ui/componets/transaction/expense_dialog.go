package transaction

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/application/validator"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
	"github.com/nelsonmarro/verith/internal/ui/componets/category"
)

// AddExpenseDialog handles the simplified UI for registering an outcome/expense.
type AddExpenseDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	recurService    RecurringTransactionService
	categoryService CategoryService
	callbackAction  func()

	// UI Components
	dateEntry         *componets.LatinDateEntry
	amountEntry       *widget.Entry
	descriptionEntry  *widget.Entry
	categoryLabel     *widget.Label
	searchCategoryBtn *widget.Button
	attachmentLabel   *widget.Label
	searchFileBtn     *widget.Button

	// Recurrence UI
	isRecurringCheck *widget.Check
	intervalSelect   *widget.Select

	// Data
	accountID        int
	selectedCategory *domain.Category
	attachmentPath   string
	currentUser      domain.User
}

func NewAddExpenseDialog(
	win fyne.Window,
	l *log.Logger,
	txs TransactionService,
	rs RecurringTransactionService,
	cs CategoryService,
	callback func(),
	accountID int,
	currentUser domain.User,
) *AddExpenseDialog {
	d := &AddExpenseDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		recurService:     rs,
		categoryService:  cs,
		callbackAction:   callback,
		accountID:        accountID,
		currentUser:      currentUser,
		dateEntry:        componets.NewLatinDateEntry(win),
		amountEntry:      widget.NewEntry(),
		descriptionEntry: widget.NewMultiLineEntry(),
		categoryLabel:    widget.NewLabel("Seleccione Categoría"),
		attachmentLabel:  widget.NewLabel("Ninguno"),
		isRecurringCheck: widget.NewCheck("Gasto Recurrente", nil),
		intervalSelect:   widget.NewSelect([]string{"Mensual", "Semanal"}, nil),
	}

	d.dateEntry.SetDate(time.Now())
	d.amountEntry.SetPlaceHolder("0.00")
	d.descriptionEntry.SetPlaceHolder("Detalle del gasto...")
	
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
		// Pre-filter ONLY OUTCOME categories
		searchDialog := category.NewCategorySearchDialog(
			d.mainWin,
			d.logger,
			d.categoryService,
			func(cat *domain.Category) {
				d.selectedCategory = cat
				d.categoryLabel.SetText(cat.Name)
			},
		)
		searchDialog.SetFilterType(domain.Outcome)
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

func (d *AddExpenseDialog) Show() {
	categoryContainer := container.NewBorder(nil, nil, nil, d.searchCategoryBtn, d.categoryLabel)
	attachmentContainer := container.NewBorder(nil, nil, nil, d.searchFileBtn, d.attachmentLabel)

	recurLabel := componets.NewHoverableLabel("¿Es Recurrente?", d.mainWin.Canvas())
	recurLabel.SetTooltip("Se creará automáticamente la siguiente transacción en el próximo periodo (mes/semana).\nTomando como base la fecha de hoy.")

	form := widget.NewForm(
		widget.NewFormItem("Fecha", d.dateEntry),
		widget.NewFormItem("Categoría", categoryContainer),
		widget.NewFormItem("Monto Total", d.amountEntry),
		widget.NewFormItem("Descripción", d.descriptionEntry),
		widget.NewFormItem("Adjunto", attachmentContainer),
		widget.NewFormItem("", container.NewBorder(nil, nil, recurLabel, nil, d.isRecurringCheck)),
		widget.NewFormItem("Frecuencia", d.intervalSelect),
	)

	dlg := dialog.NewCustomConfirm("Registrar Gasto", "Guardar", "Cancelar", 
		container.NewVScroll(container.NewPadded(form)), 
		func(confirm bool) {
			if confirm {
				d.submit()
			}
		}, d.mainWin)
	
	dlg.Resize(fyne.NewSize(500, 500))
	dlg.Show()
}

func (d *AddExpenseDialog) submit() {
	// 1. Validation
	amount, err := strconv.ParseFloat(d.amountEntry.Text, 64)
	if err != nil || amount <= 0 {
		dialog.ShowError(fmt.Errorf("monto inválido"), d.mainWin)
		return
	}

	if d.dateEntry.Date == nil {
		dialog.ShowError(fmt.Errorf("fecha inválida"), d.mainWin)
		return
	}
	txDate := *d.dateEntry.Date

	if validator.IsDateAfter(txDate, time.Now()) {
		dialog.ShowError(fmt.Errorf("no puede registrar gastos en el futuro"), d.mainWin)
		return
	}

	if d.selectedCategory == nil {
		dialog.ShowError(fmt.Errorf("seleccione una categoría"), d.mainWin)
		return
	}

	if d.descriptionEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("ingrese una descripción"), d.mainWin)
		return
	}

	// 2. Build Objects
	var attachment *string
	if d.attachmentPath != "" {
		attachment = &d.attachmentPath
	}

	// Transaction Item (One single item representing the expense)
	item := domain.TransactionItem{
		Description: d.descriptionEntry.Text,
		Quantity:    1,
		UnitPrice:   amount,
		TaxRate:     0, // Expenses usually recorded as total (or 0% tax for simplicity unless detailed)
		Subtotal:    amount,
	}

	tx := &domain.Transaction{
		Description:     d.descriptionEntry.Text,
		Amount:          amount,
		TransactionDate: txDate,
		AccountID:       d.accountID,
		CategoryID:      d.selectedCategory.ID,
		AttachmentPath:  attachment,
		Subtotal0:       amount, // Assuming expenses go to 0% bucket for internal record or just Total
		TaxAmount:       0,
		Items:           []domain.TransactionItem{item},
		// No TaxPayer for simple expense
	}

	// 3. Save
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := d.txService.CreateTransaction(ctx, tx, d.currentUser); err != nil {
		dialog.ShowError(fmt.Errorf("error guardando gasto: %w", err), d.mainWin)
		return
	}

	// 4. Recurrence
	if d.isRecurringCheck.Checked {
		interval := domain.IntervalMonthly
		if d.intervalSelect.Selected == "Semanal" {
			interval = domain.IntervalWeekly
		}
		
		// Calculate Next Run Date: Start Date + Interval
		// Because we just created the "current" transaction manually.
		nextRun := txDate
		if interval == domain.IntervalMonthly {
			nextRun = nextRun.AddDate(0, 1, 0)
		} else {
			nextRun = nextRun.AddDate(0, 0, 7)
		}

		rt := &domain.RecurringTransaction{
			AccountID:   d.accountID,
			CategoryID:  d.selectedCategory.ID,
			Description: d.descriptionEntry.Text,
			Amount:      amount,
			StartDate:   txDate,
			NextRunDate: nextRun, // Important!
			Interval:    interval,
			IsActive:    true,
		}
		if err := d.recurService.Create(ctx, rt); err != nil {
			d.logger.Printf("Error creating recurrence: %v", err)
		}
	}

	dialog.ShowInformation("Éxito", "Gasto registrado correctamente", d.mainWin)
	d.callbackAction()
}
