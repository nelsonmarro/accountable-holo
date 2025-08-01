package transaction

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/category"
	"github.com/shopspring/decimal"
)

// AdjustmentDialogHandler handles the logic for the adjustment transaction dialog.
type AdjustmentDialogHandler struct {
	dialog      dialog.Dialog
	txService   TransactionService
	catService  CategoryService
	errorLogger *log.Logger
	parent      fyne.Window
	onConfirm   func()

	// UI Components
	descriptionEntry  *widget.Entry
	amountLabel       *widget.Label
	dateLabel         *widget.Label
	categoryLabel     *widget.Label
	searchCategoryBtn *widget.Button

	// Data
	accountID        int
	selectedCategory *domain.Category
	amount           decimal.Decimal
	transactionDate  time.Time
}

// NewAdjustmentTransactionDialog creates a new dialog for creating a reconciliation adjustment transaction.
func NewAdjustmentTransactionDialog(
	parent fyne.Window,
	errorLogger *log.Logger,
	txService TransactionService,
	catService CategoryService,
	reconciliationData *domain.Reconciliation,
	onConfirm func(),
) *AdjustmentDialogHandler {
	h := &AdjustmentDialogHandler{
		parent:           parent,
		errorLogger:      errorLogger,
		txService:        txService,
		catService:       catService,
		onConfirm:        onConfirm,
		accountID:        reconciliationData.AccountID,
		descriptionEntry: widget.NewMultiLineEntry(),
		amountLabel:      widget.NewLabel(""),
		dateLabel:        widget.NewLabel(""),
		categoryLabel:    widget.NewLabel("Buscando categoría..."),
		transactionDate:  time.Now(),
	}

	h.searchCategoryBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchDialog := category.NewCategorySearchDialog(
			h.parent,
			h.errorLogger,
			h.catService,
			func(cat *domain.Category) {
				h.selectedCategory = cat
				h.categoryLabel.SetText(cat.Name)
			},
		)
		searchDialog.Show()
	})

	h.prefillForm(reconciliationData)

	categoryContainer := container.NewBorder(nil, nil, nil, h.searchCategoryBtn, h.categoryLabel)

	formItems := []*widget.FormItem{
		{Text: "Descripción", Widget: h.descriptionEntry},
		{Text: "Monto", Widget: h.amountLabel},
		{Text: "Fecha", Widget: h.dateLabel},
		{Text: "Categoría", Widget: categoryContainer},
	}

	h.dialog = dialog.NewForm("Crear Transacción de Ajuste", "Crear", "Cancelar", formItems, h.submit, parent)
	h.dialog.Resize(fyne.NewSize(500, 350))
	return h
}

// prefillForm sets the initial values for the adjustment transaction.
func (h *AdjustmentDialogHandler) prefillForm(data *domain.Reconciliation) {
	h.amount = data.Difference.Abs()
	h.amountLabel.SetText(fmt.Sprintf("$%s", h.amount.StringFixed(2)))

	description := fmt.Sprintf("Ajuste por reconciliación de cuenta al %s.", data.EndDate.Format("2006-01-02"))
	h.descriptionEntry.SetText(description)

	h.dateLabel.SetText(h.transactionDate.Format("01/02/2006"))

	go h.findAndSetCategory(domain.Adjustment, "Ajuste por Reconciliación")

	h.searchCategoryBtn.Disable()
}

func (h *AdjustmentDialogHandler) findAndSetCategory(catType domain.CategoryType, catName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	category, err := h.catService.GetCategoryByTypeAndName(ctx, catType, catName)
	if err != nil {
		fyne.Do(func() { h.categoryLabel.SetText("Error: Categoría no encontrada") })
		return
	}

	h.selectedCategory = category
	fyne.Do(func() {
		h.categoryLabel.SetText(category.Name)
	})
}

// submit handles the form submission.
func (h *AdjustmentDialogHandler) submit(confirmed bool) {
	if !confirmed {
		return
	}

	if h.selectedCategory == nil {
		dialog.ShowError(fmt.Errorf("la categoría de ajuste no se ha podido cargar"), h.parent)
		return
	}

	progressDialog := dialog.NewCustomWithoutButtons("Creando ajuste...", widget.NewProgressBarInfinite(), h.parent)
	progressDialog.Show()

	go func() {
		amountFloat, _ := h.amount.Float64()

		tx := &domain.Transaction{
			AccountID:       h.accountID,
			CategoryID:      h.selectedCategory.ID,
			Description:     h.descriptionEntry.Text,
			Amount:          amountFloat,
			TransactionDate: h.transactionDate,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := h.txService.CreateTransaction(ctx, tx)
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
			h.dialog.Hide()
		})
	}()
}

// Show displays the dialog.
func (h *AdjustmentDialogHandler) Show() {
	h.dialog.Show()
}