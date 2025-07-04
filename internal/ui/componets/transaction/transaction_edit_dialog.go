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

// EditTransactionDialog holds the state and logic for the 'Edit Transaction' dialog.
type EditTransactionDialog struct {
	mainWin         fyne.Window
	logger          *log.Logger
	txService       TransactionService
	categoryService CategoryService
	callbackAction  func()
	txID            int

	// UI Components
	descriptionEntry *widget.Entry
	amountEntry      *widget.Entry
	dateEntry        *widget.Entry
	categorySelect   *widget.SelectEntry

	// Data
	accountID  int
	categories []domain.Category
}

// NewEditTransactionDialog creates a new dialog handler for the edit action.
func NewEditTransactionDialog(win fyne.Window, l *log.Logger, txs TransactionService, cs CategoryService, callback func(), txID int, accountID int) *EditTransactionDialog {
	return &EditTransactionDialog{
		mainWin:          win,
		logger:           l,
		txService:        txs,
		categoryService:  cs,
		callbackAction:   callback,
		txID:             txID,
		accountID:        accountID,
		descriptionEntry: widget.NewEntry(),
		amountEntry:      widget.NewEntry(),
		dateEntry:        widget.NewEntry(),
		categorySelect:   widget.NewSelectEntry([]string{}),
	}
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
			dialog.ShowError(fmt.Errorf("Error fetching transaction: %w", err), d.mainWin)
		})
	}

	d.fetchTransaction(onSuccess, onFailure)
}

func (d *EditTransactionDialog) fetchTransaction(onSuccess func(tx *domain.Transaction), onFailure func(err error)) {
	progress := dialog.NewCustomWithoutButtons("Loading Transaction...", widget.NewProgressBarInfinite(), d.mainWin)
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

		categories, err := d.categoryService.GetAllCategories(ctx)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
			})
			onFailure(err)
			return
		}
		d.categories = categories

		onSuccess(tx)
	}()
}

func (d *EditTransactionDialog) showEditForm(tx *domain.Transaction) {
	d.descriptionEntry.SetText(tx.Description)
	d.amountEntry.SetText(fmt.Sprintf("%.2f", tx.Amount))
	d.dateEntry.SetText(tx.TransactionDate.Format("2006-01-02"))

	categoryNames := make([]string, len(d.categories))
	for i, cat := range d.categories {
		categoryNames[i] = cat.Name
	}
	d.categorySelect.SetOptions(categoryNames)

	for _, cat := range d.categories {
		if cat.ID == tx.CategoryID {
			d.categorySelect.SetText(cat.Name)
			break
		}
	}

	formDialog := dialog.NewForm("Edit Transaction", "Save", "Cancel",
		TransactionForm(
			d.descriptionEntry,
			d.amountEntry,
			d.dateEntry,
			d.categorySelect,
		),
		d.handleSubmit, // The submit callback
		d.mainWin,
	)
	formDialog.Resize(fyne.NewSize(480, 300))
	formDialog.Show()
}

func (d *EditTransactionDialog) handleSubmit(valid bool) {
	if !valid {
		return
	}

	progress := dialog.NewCustomWithoutButtons("Saving Changes...", widget.NewProgressBarInfinite(), d.mainWin)
	progress.Show()

	go func() {
		amount, _ := strconv.ParseFloat(d.amountEntry.Text, 64)
		transactionDate, _ := time.Parse("2006-01-02", d.dateEntry.Text)

		var categoryID int
		for _, cat := range d.categories {
			if cat.Name == d.categorySelect.SelectedText() {
				categoryID = cat.ID
				break
			}
		}

		updatedTx := &domain.Transaction{
			BaseEntity:      domain.BaseEntity{ID: d.txID},
			Description:     d.descriptionEntry.Text,
			Amount:          amount,
			TransactionDate: transactionDate,
			AccountID:       d.accountID,
			CategoryID:      categoryID,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := d.txService.UpdateTransaction(ctx, updatedTx)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(err, d.mainWin)
			})
			d.logger.Printf("Error updating transaction %d: %v", d.txID, err)
			return
		}

		fyne.Do(func() {
			progress.Hide()
			dialog.ShowInformation("Success", "Changes have been saved.", d.mainWin)
		})

		go d.callbackAction()
	}()
}
