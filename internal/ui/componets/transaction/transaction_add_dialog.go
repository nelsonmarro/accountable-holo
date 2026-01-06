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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/category"
)

// AddTransactionDialog holds the state and logic for the 'Add Transaction' dialog.
type AddTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	recurService    RecurringTransactionService // New dependency
	categoryService CategoryService
	callbackAction  func()

	// UI Components
	descriptionEntry  *widget.Entry
	amountEntry       *widget.Entry
	dateEntry         *widget.DateEntry
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

// NewAddTransactionDialog creates a new dialog handler.
func NewAddTransactionDialog(
	win fyne.Window,
	l *log.Logger,
	txs TransactionService,
	rs RecurringTransactionService, // New dependency
	cs CategoryService,
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
		callbackAction:   callback,
		descriptionEntry: widget.NewMultiLineEntry(),
		amountEntry:      widget.NewEntry(),
		dateEntry:        widget.NewDateEntry(),
		accountID:        accountID,
		categoryLabel:    widget.NewLabel("Ninguna seleccionada"),
		attachmentLabel:  widget.NewLabel("Ninguno"),
		currentUser:      currentUser,
		isRecurringCheck: widget.NewCheck("", nil),
		intervalSelect:   widget.NewSelect([]string{"Mensual", "Semanal"}, nil),
	}
	d.dateEntry.SetText(time.Now().Format("01/02/2006"))
	d.intervalSelect.SetSelected("Mensual")
	d.intervalSelect.Hide() // Hidden by default

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

// Show creates and displays the Fyne form dialog.
func (d *AddTransactionDialog) Show() {
	categoryContainer := container.NewBorder(nil, nil, nil, d.searchCategoryBtn, d.categoryLabel)
	attachmentContainer := container.NewBorder(nil, nil, nil, d.searchFileBtn, d.attachmentLabel)

	recurLabel := componets.NewHoverableLabel("¿Es Recurrente?", d.mainWin.Canvas())
	recurLabel.SetTooltip("Crea esta transacción automáticamente cada periodo (mes/semana).\nSe calculara a partir de la fecha de la transacción.")

	formDialog := dialog.NewForm("Crear Transacción", "Guardar", "Cancelar",
		TransactionForm(
			d.descriptionEntry,
			d.amountEntry,
			d.dateEntry,
			categoryContainer,
			attachmentContainer,
			d.isRecurringCheck,
			d.intervalSelect,
			recurLabel,
		),
		d.handleSubmit, // Pass the method as the callback
		d.mainWin,
	)
	formDialog.Resize(fyne.NewSize(560, 450)) // Slightly taller
	formDialog.Show()
}

func (d *AddTransactionDialog) handleSubmit(valid bool) {
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

	if d.selectedCategory == nil {
		dialog.ShowError(fmt.Errorf("por favor, seleccione una categoría"), d.mainWin)
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Espere...", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		amount, _ := strconv.ParseFloat(d.amountEntry.Text, 64)

		var attachmentPathPtr *string
		if d.attachmentPath != "" {
			attachmentPathPtr = &d.attachmentPath
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// 1. Create the base transaction (Always created immediately)
		tx := &domain.Transaction{
			Description:     d.descriptionEntry.Text,
			Amount:          amount,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      d.selectedCategory.ID,
			AttachmentPath:  attachmentPathPtr,
		}

		err := d.txService.CreateTransaction(ctx, tx, d.currentUser)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al crear la Transacción: %w", err), d.mainWin)
			})
			return
		}

		// 2. If Recurring, create the recurrence rule
		if d.isRecurringCheck.Checked {
			interval := domain.IntervalMonthly
			if d.intervalSelect.Selected == "Semanal" {
				interval = domain.IntervalWeekly
			}

			// Calculate NEXT run date based on the transaction date just entered
			nextRun := transactionDate
			if interval == domain.IntervalMonthly {
				nextRun = nextRun.AddDate(0, 1, 0)
			} else {
				nextRun = nextRun.AddDate(0, 0, 7)
			}

			rt := &domain.RecurringTransaction{
				Description: d.descriptionEntry.Text,
				Amount:      amount,
				AccountID:   d.accountID,
				CategoryID:  d.selectedCategory.ID,
				Interval:    interval,
				StartDate:   transactionDate,
				NextRunDate: nextRun,
				IsActive:    true,
			}

			if err := d.recurService.Create(ctx, rt); err != nil {
				// Non-fatal error, but warn the user
				d.logger.Printf("Failed to create recurrence: %v", err)
				fyne.Do(func() {
					dialog.ShowInformation("Advertencia", "La transacción se guardó, pero hubo un error al programar la recurrencia.", d.mainWin)
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
