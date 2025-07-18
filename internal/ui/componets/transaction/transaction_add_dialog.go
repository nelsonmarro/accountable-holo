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
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/category"
)

// AddTransactionDialog holds the state and logic for the 'Add Transaction' dialog.
type AddTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
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

	// Data
	accountID        int
	categories       []domain.Category
	selectedCategory *domain.Category
	attachmentPath   string
}

// NewAddTransactionDialog creates a new dialog handler.
func NewAddTransactionDialog(
	win fyne.Window,
	l *log.Logger,
	txs TransactionService,
	cs CategoryService,
	callback func(),
	accountID int,
) *AddTransactionDialog {
	d := &AddTransactionDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		categoryService:  cs,
		callbackAction:   callback,
		descriptionEntry: widget.NewMultiLineEntry(),
		amountEntry:      widget.NewEntry(),
		dateEntry:        widget.NewDateEntry(),
		accountID:        accountID,
		categoryLabel:    widget.NewLabel("Ninguna seleccionada"),
		attachmentLabel:  widget.NewLabel("Ninguno"),
	}
	d.dateEntry.SetText(time.Now().Format("01/02/2006"))

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
	formDialog := dialog.NewForm("Crear Transacción", "Guardar", "Cancelar",
		TransactionForm(
			d.descriptionEntry,
			d.amountEntry,
			d.dateEntry,
			categoryContainer,
			attachmentContainer,
		),
		d.handleSubmit, // Pass the method as the callback
		d.mainWin,
	)

	formDialog.Resize(fyne.NewSize(560, 400))
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

		tx := &domain.Transaction{
			Description:     d.descriptionEntry.Text,
			Amount:          amount,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      d.selectedCategory.ID,
			AttachmentPath:  d.attachmentPath,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.txService.CreateTransaction(ctx, tx)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al crear la Transacción: %w", err), d.mainWin)
			})
			d.logger.Println("Error creating transaction:", err)
			return
		}

		fyne.Do(func() {
			progressDialog.Hide()
			dialog.ShowInformation("Transacción Creada", "Transacción creada exitosamente!", d.mainWin)
		})

		go d.callbackAction()
	}()
}