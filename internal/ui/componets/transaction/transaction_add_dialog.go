package transaction

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// AddTransactionDialog holds the state and logic for the 'Add Transaction' dialog.
type AddTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	categoryService CategoryService
	callbackAction  func()

	// UI Components
	descriptionEntry *widget.Entry
	amountEntry      *widget.Entry
	dateEntry        *widget.Entry
	categorySelect   *widget.SelectEntry

	// Data
	accountID  int
	categories []domain.Category
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
	return &AddTransactionDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		categoryService:  cs,
		callbackAction:   callback,
		descriptionEntry: widget.NewEntry(),
		amountEntry:      widget.NewEntry(),
		dateEntry:        widget.NewEntry(),
		accountID:        accountID,
		categorySelect:   widget.NewSelectEntry([]string{}),
	}
}

// Show creates and displays the Fyne form dialog.
func (d *AddTransactionDialog) Show() {
	d.loadData()

	formDialog := dialog.NewForm("Create Transaction", "Save", "Cancel",
		TransactionForm(
			d.descriptionEntry,
			d.amountEntry,
			d.dateEntry,
			d.categorySelect,
		),
		d.handleSubmit, // Pass the method as the callback
		d.mainWin,
	)

	formDialog.Resize(fyne.NewSize(480, 300))
	formDialog.Show()
}

func (d *AddTransactionDialog) loadData() {
	go func() {
		progressDialog := dialog.NewCustomWithoutButtons("Cargando...", widget.NewProgressBarInfinite(), d.mainWin)
		fyne.Do(func() {
			progressDialog.Show()
		})

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		categories, err := d.categoryService.GetAllCategories(ctx)
		if err != nil {
			d.logger.Println("Error loading categories:", err)
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al cargar las categorias: %w", err), d.mainWin)
			})
			return
		}
		d.categories = categories

		categoryNames := make([]string, len(categories))
		for i, cat := range categories {
			categoryNames[i] = cat.Name
		}

		fyne.Do(func() {
			progressDialog.Hide()
			d.categorySelect.SetOptions(categoryNames)
			d.categorySelect.SetText(categoryNames[0])
			d.categorySelect.Refresh()
		})
	}()
}

func (d *AddTransactionDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Please wait", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		amount, _ := strconv.ParseFloat(d.amountEntry.Text, 64)
		transactionDate, _ := time.Parse("2006-01-02", d.dateEntry.Text)

		var categoryID int
		for _, cat := range d.categories {
			if cat.Name == d.categorySelect.Text {
				categoryID = cat.ID
				break
			}
		}

		tx := &domain.Transaction{
			Description:     d.descriptionEntry.Text,
			Amount:          amount,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      categoryID,
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
