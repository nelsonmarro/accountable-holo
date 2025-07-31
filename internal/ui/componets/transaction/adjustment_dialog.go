package transaction

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

// AdjustmentDialogHandler handles the logic for the adjustment transaction dialog.
type AdjustmentDialogHandler struct {
	dialog      dialog.Dialog
	form        *TransactionForm
	txService   TransactionService
	catService  CategoryService
	errorLogger *log.Logger
	parent      fyne.Window
	onConfirm   func()
}

// NewAdjustmentTransactionDialog creates a new dialog for creating a reconciliation adjustment transaction.
func NewAdjustmentTransactionDialog(
	parent fyne.Window,
	errorLogger *log.Logger,
	txService TransactionService,
	catService CategoryService,
	onConfirm func(),
	reconciliationData *domain.Reconciliation,
) *AdjustmentDialogHandler {
	h := &AdjustmentDialogHandler{
		parent:      parent,
		errorLogger: errorLogger,
		txService:   txService,
		catService:  catService,
		onConfirm:   onConfirm,
	}

	h.form = NewTransactionForm(parent, errorLogger, catService, reconciliationData.AccountID)
	h.prefillForm(reconciliationData)

	h.dialog = dialog.NewForm("Crear Transacción de Ajuste", "Crear", "Cancelar", h.form.GetFormItems(), h.submit, parent)
	h.dialog.Resize(fyne.NewSize(400, 300))
	return h
}

// prefillForm sets the initial values for the adjustment transaction.
func (h *AdjustmentDialogHandler) prefillForm(data *domain.Reconciliation) {
	amount := data.Difference.Abs()
	h.form.Amount.SetText(amount.StringFixed(2))
	h.form.Amount.Disable()

	description := fmt.Sprintf("Ajuste por reconciliación de cuenta al %s.", data.EndDate.Format("2006-01-02"))
	h.form.Description.SetText(description)

	// Assuming h.form.TransactionDate is a *widget.DateEntry
	h.form.TransactionDate.SetText(data.EndDate.Format("2006-01-02"))

	// Determine if it's an income or outcome adjustment
	var catType domain.CategoryType
	if data.Difference.IsPositive() {
		catType = domain.Income
	} else {
		catType = domain.Outcome
	}

	// Find and set the system category for adjustments
	go h.form.SetCategoryByTypeAndName(catType, "Ajuste") // Assuming "Ajuste" is the name for adjustment categories

	h.form.Account.Disable()
	h.form.Category.Disable()
}

// submit handles the form submission.
func (h *AdjustmentDialogHandler) submit(confirmed bool) {
	if !confirmed {
		return
	}

	if err := h.form.Validate(); err != nil {
		dialog.ShowError(err, h.parent)
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Espere...", widget.NewProgressBarInfinite(), h.parent)
	progressDialog.Show()

	go func() {
		amount, _ := decimal.NewFromString(h.form.Amount.Text)
		selectedCat, err := h.form.Category.GetSelected()
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al obtener la categoría: %w", err), h.parent)
			})
			return
		}

		transactionDate, err := time.Parse("2006-01-02", h.form.TransactionDate.Text)
		if err != nil {
			fyne.Do(func() {

				dialog.ShowError(fmt.Errorf("formato de fecha inválido: %w", err), h.parent)
			})
			return
		}

		tx := &domain.Transaction{
			AccountID:       h.form.Account.GetSelected(),
			CategoryID:      selectedCat.ID,
			Description:     h.form.Description.Text,
			Amount:          math.Abs(amount.InexactFloat64()),
			TransactionDate: transactionDate,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err = h.txService.CreateTransaction(ctx, tx)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al crear la transacción de ajuste: %w", err), h.parent)
			})
			h.errorLogger.Println("Error creating adjustment transaction:", err)
			return
		}

		fyne.Do(func() {
			progressDialog.Hide()
			dialog.ShowInformation("Ajuste Creado", "Transacción de ajuste creada exitosamente!", h.parent)
			if h.onConfirm != nil {
				h.onConfirm()
			}
		})
		fyne.Do(func() { h.dialog.Hide() })
	}()
}

// Show displays the dialog.
func (h *AdjustmentDialogHandler) Show() {
	h.dialog.Show()
}