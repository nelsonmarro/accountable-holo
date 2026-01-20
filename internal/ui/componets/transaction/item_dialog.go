package transaction

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
)

type ItemDialog struct {
	parent fyne.Window
	onSave func(domain.TransactionItem)

	// Widgets
	descEntry  *widget.Entry
	qtyEntry   *widget.Entry
	priceEntry *widget.Entry
	taxSelect  *widget.Select
}

func NewItemDialog(parent fyne.Window, onSave func(domain.TransactionItem)) *ItemDialog {
	return &ItemDialog{
		parent:     parent,
		onSave:     onSave,
		descEntry:  widget.NewEntry(),
		qtyEntry:   widget.NewEntry(),
		priceEntry: widget.NewEntry(),
		taxSelect:  widget.NewSelect([]string{"IVA 15%", "IVA 0%", "Exento"}, nil),
	}
}

func (d *ItemDialog) Show() {
	d.descEntry.SetPlaceHolder("Descripción del producto/servicio")
	d.qtyEntry.SetText("1")
	d.priceEntry.SetPlaceHolder("0.00")
	d.taxSelect.SetSelected("IVA 15%")

	form := widget.NewForm(
		widget.NewFormItem("Descripción", d.descEntry),
		widget.NewFormItem("Cantidad", d.qtyEntry),
		widget.NewFormItem("Precio Unitario", d.priceEntry),
		widget.NewFormItem("Impuesto", d.taxSelect),
	)

	dlg := dialog.NewCustomConfirm("Nuevo Ítem", "Agregar", "Cancelar", form, func(confirm bool) {
		if !confirm {
			return
		}

		// Validar y procesar
		qty, _ := strconv.ParseFloat(d.qtyEntry.Text, 64)
		price, _ := strconv.ParseFloat(d.priceEntry.Text, 64)

		if d.descEntry.Text == "" || qty <= 0 || price < 0 {
			dialog.ShowError(fmt.Errorf("por favor ingrese datos válidos"), d.parent)
			return
		}

		taxRate := 0
		if d.taxSelect.Selected == "IVA 15%" {
			taxRate = 4 // Código SRI para 15%
		}

		item := domain.TransactionItem{
			Description: d.descEntry.Text,
			Quantity:    qty,
			UnitPrice:   price,
			TaxRate:     taxRate,
			Subtotal:    qty * price,
		}

		d.onSave(item)
	}, d.parent)

	// Force a comfortable size
	dlg.Resize(fyne.NewSize(400, 300))
	dlg.Show()
}
