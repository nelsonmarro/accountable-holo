package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
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

// EditExpenseDialog handles the simplified UI for editing an outcome/expense.
type EditExpenseDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	recurService    RecurringTransactionService
	categoryService CategoryService
	callbackAction  func()
	txID            int

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

func NewEditExpenseDialog(
	win fyne.Window,
	l *log.Logger,
	txs TransactionService,
	rs RecurringTransactionService,
	cs CategoryService,
	callback func(),
	txID int,
	accountID int,
	currentUser domain.User,
) *EditExpenseDialog {
	return &EditExpenseDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		recurService:     rs,
		categoryService:  cs,
		callbackAction:   callback,
		txID:             txID,
		accountID:        accountID,
		currentUser:      currentUser,
		dateEntry:        componets.NewLatinDateEntry(win),
		amountEntry:      widget.NewEntry(),
		descriptionEntry: widget.NewMultiLineEntry(),
		categoryLabel:    widget.NewLabel("Cargando..."),
		attachmentLabel:  widget.NewLabel("Ninguno"),
		isRecurringCheck: widget.NewCheck("Gasto Recurrente", nil),
		intervalSelect:   widget.NewSelect([]string{"Mensual", "Semanal"}, nil),
	}
}

func (d *EditExpenseDialog) Show() {
	progress := dialog.NewCustomWithoutButtons("Cargando...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tx, err := d.txService.GetTransactionByID(ctx, d.txID)
		if err != nil || tx == nil {
			fyne.Do(func() {
				progress.Hide()
				errMsg := "error al cargar gasto"
				if err != nil {
					errMsg = fmt.Sprintf("%s: %v", errMsg, err)
				}
				dialog.ShowError(errors.New(errMsg), d.mainWin)
			})
			return
		}

		cat, _ := d.categoryService.GetCategoryByID(ctx, tx.CategoryID)

		// Check if it's recurring (Heuristic search)
		isCurrentlyRecurring := false
		recurRules, _ := d.recurService.GetAll(ctx)
		for _, rule := range recurRules {
			// Heuristic: Match Account, Category and Description (handling the Recurrente suffix)
			descMatch := tx.Description == rule.Description ||
				tx.Description == fmt.Sprintf("%s (Recurrente)", rule.Description) ||
				(rule.Description != "" && strings.HasPrefix(tx.Description, rule.Description))

			if rule.IsActive &&
				rule.AccountID == tx.AccountID &&
				rule.CategoryID == tx.CategoryID &&
				descMatch {
				isCurrentlyRecurring = true
				break
			}
		}

		fyne.Do(func() {
			d.dateEntry.SetDate(tx.TransactionDate)
			d.amountEntry.SetText(fmt.Sprintf("%.2f", tx.Amount))
			d.descriptionEntry.SetText(tx.Description)
			if cat != nil {
				d.selectedCategory = cat
				d.categoryLabel.SetText(cat.Name)
			}

			d.isRecurringCheck.SetChecked(isCurrentlyRecurring)

			if tx.AttachmentPath != nil && *tx.AttachmentPath != "" {
				d.attachmentPath = *tx.AttachmentPath
				d.attachmentLabel.SetText("Archivo adjunto guardado")
			}

			d.intervalSelect.SetSelected("Mensual")
			if isCurrentlyRecurring {
				d.intervalSelect.Show()
			} else {
				d.intervalSelect.Hide()
			}

			d.isRecurringCheck.OnChanged = func(checked bool) {
				if checked {
					d.intervalSelect.Show()
				} else {
					d.intervalSelect.Hide()
				}
			}

			d.searchCategoryBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
				searchDialog := category.NewCategorySearchDialog(d.mainWin, d.logger, d.categoryService, func(c *domain.Category) {
					d.selectedCategory = c
					d.categoryLabel.SetText(c.Name)
				})
				searchDialog.SetFilterType(domain.Outcome)
				searchDialog.Show()
			})

			d.searchFileBtn = widget.NewButtonWithIcon("", theme.FileIcon(), func() {
				fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err == nil && reader != nil {
						d.attachmentPath = reader.URI().Path()
						d.attachmentLabel.SetText(reader.URI().Name())
					}
				}, d.mainWin)
				fileDialog.Show()
			})

			progress.Hide()
			d.showForm()
		})
	}()
}

func (d *EditExpenseDialog) showForm() {
	categoryContainer := container.NewBorder(nil, nil, nil, d.searchCategoryBtn, d.categoryLabel)
	attachmentContainer := container.NewBorder(nil, nil, nil, d.searchFileBtn, d.attachmentLabel)

	recurLabel := componets.NewHoverableLabel("¿Es Recurrente?", d.mainWin.Canvas())
	recurLabel.SetTooltip("Crea una nueva regla de automatización a partir de esta transacción.")

	form := widget.NewForm(
		widget.NewFormItem("Fecha", d.dateEntry),
		widget.NewFormItem("Categoría", categoryContainer),
		widget.NewFormItem("Monto Total", d.amountEntry),
		widget.NewFormItem("Descripción", d.descriptionEntry),
		widget.NewFormItem("Adjunto", attachmentContainer),
		widget.NewFormItem("", container.NewBorder(nil, nil, recurLabel, nil, d.isRecurringCheck)),
		widget.NewFormItem("Frecuencia", d.intervalSelect),
	)

	dlg := dialog.NewCustomConfirm("Editar Gasto", "Guardar Cambios", "Cancelar",
		container.NewVScroll(container.NewPadded(form)),
		func(confirm bool) {
			if confirm {
				d.submit()
			}
		}, d.mainWin)

	dlg.Resize(fyne.NewSize(500, 500))
	dlg.Show()
}

func (d *EditExpenseDialog) submit() {
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
		dialog.ShowError(fmt.Errorf("fecha no puede ser futura"), d.mainWin)
		return
	}

	var attachment *string
	if d.attachmentPath != "" {
		attachment = &d.attachmentPath
	}

	// Simplification: Update current tx with single item
	item := domain.TransactionItem{
		Description: d.descriptionEntry.Text,
		Quantity:    1,
		UnitPrice:   amount,
		TaxRate:     0,
		Subtotal:    amount,
	}

	updatedTx := &domain.Transaction{
		BaseEntity:      domain.BaseEntity{ID: d.txID},
		Description:     d.descriptionEntry.Text,
		Amount:          amount,
		TransactionDate: txDate,
		AccountID:       d.accountID,
		CategoryID:      d.selectedCategory.ID,
		AttachmentPath:  attachment,
		Subtotal0:       amount,
		TaxAmount:       0,
		Items:           []domain.TransactionItem{item},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := d.txService.UpdateTransaction(ctx, updatedTx, d.currentUser); err != nil {
		dialog.ShowError(fmt.Errorf("error al actualizar gasto: %w", err), d.mainWin)
		return
	}

	// Recurrence logic (same as create)
	if d.isRecurringCheck.Checked {
		interval := domain.IntervalMonthly
		if d.intervalSelect.Selected == "Semanal" {
			interval = domain.IntervalWeekly
		}
		rt := &domain.RecurringTransaction{
			AccountID:   d.accountID,
			CategoryID:  d.selectedCategory.ID,
			Description: d.descriptionEntry.Text,
			Amount:      amount,
			StartDate:   txDate,
			NextRunDate: txDate.AddDate(0, 1, 0), // Basic heuristic
			Interval:    interval,
			IsActive:    true,
		}
		_ = d.recurService.Create(ctx, rt)
	}

	dialog.ShowInformation("Éxito", "Gasto actualizado correctamente", d.mainWin)
	d.callbackAction()
}
