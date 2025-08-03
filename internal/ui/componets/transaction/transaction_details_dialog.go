package transaction

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// DetailsDialog holds the state for the transaction details dialog.
type DetailsDialog struct {
	parent fyne.Window
	tx     *domain.Transaction
}

// NewDetailsDialog creates a new dialog handler for displaying transaction details.
func NewDetailsDialog(parent fyne.Window, tx *domain.Transaction) *DetailsDialog {
	return &DetailsDialog{
		parent: parent,
		tx:     tx,
	}
}

// Show creates and displays the dialog.
func (d *DetailsDialog) Show() {
	content := d.buildContent()
	dlg := dialog.NewCustom("Detalles de la Transacción", "Cerrar", content, d.parent)
	dlg.Resize(fyne.NewSize(700, 500))
	dlg.Show()
}

func (d *DetailsDialog) buildContent() fyne.CanvasObject {
	// Create labels for each field
	txNumberLabel := widget.NewLabel(d.tx.TransactionNumber)
	dateLabel := widget.NewLabel(d.tx.TransactionDate.Format("01/02/2006"))

	descriptionLabel := widget.NewLabel(d.tx.Description)
	descriptionLabel.Wrapping = fyne.TextWrapBreak

	categoryLabel := widget.NewLabel(d.tx.Category.Name)
	typeLabel := widget.NewLabel(string(d.tx.Category.Type))
	amountLabel := widget.NewLabel(fmt.Sprintf("$%.2f", d.tx.Amount))

	// Create a grid layout for the key-value pairs
	grid := container.New(
		layout.NewFormLayout(),
		widget.NewLabel("Número:"),
		txNumberLabel,
		widget.NewLabel("Fecha:"),
		dateLabel,
		widget.NewLabel("Descripción:"),
		descriptionLabel,
		widget.NewLabel("Categoría:"),
		categoryLabel,
		widget.NewLabel("Tipo:"),
		typeLabel,
		widget.NewLabel("Monto:"),
		amountLabel,
	)

	return container.NewPadded(grid)
}
