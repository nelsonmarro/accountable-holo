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
	descEntry      *widget.Entry
	qtyEntry       *widget.Entry
	priceEntry     *widget.Entry
	taxSelect      *widget.Select
	defaultTaxRate int
}

func NewItemDialog(parent fyne.Window, onSave func(domain.TransactionItem), defaultTax int) *ItemDialog {
	return &ItemDialog{
		parent:         parent,
		onSave:         onSave,
		descEntry:      widget.NewEntry(),
		qtyEntry:       widget.NewEntry(),
		priceEntry:     widget.NewEntry(),
		taxSelect:      widget.NewSelect([]string{"IVA 15%", "IVA 13%", "IVA 8%", "IVA 5%", "IVA 0%", "No Objeto (6)", "Exento (7)"}, nil),
		defaultTaxRate: defaultTax,
	}
}

func (d *ItemDialog) Show() {
	d.configureWidgets()

	dlg := dialog.NewCustomConfirm("Nuevo Ítem", "Agregar", "Cancelar", d.buildForm(), func(confirm bool) {
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
		switch d.taxSelect.Selected {
		case "IVA 15%":
			taxRate = 4
		case "IVA 13%":
			taxRate = 10
		case "IVA 8%":
			taxRate = 8
		case "IVA 5%":
			taxRate = 5
		case "IVA 0%":
			taxRate = 0
		case "No Objeto (6)":
			taxRate = 6
		case "Exento (7)":
			taxRate = 7
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

func (d *ItemDialog) buildForm() *widget.Form {
	items := []*widget.FormItem{
		widget.NewFormItem("Descripción", d.descEntry),
		widget.NewFormItem("Cantidad", d.qtyEntry),
		widget.NewFormItem("Precio Unitario", d.priceEntry),
		widget.NewFormItem("Impuesto", d.taxSelect),
	}

	// Solo mostrar check si hay un IVA por defecto configurado (diferente de -1)
	if d.defaultTaxRate != -1 {
		manualTaxCheck := widget.NewCheck("Cambiar IVA manual", func(checked bool) {
			if checked {
				d.taxSelect.Enable()
			} else {
				d.taxSelect.Disable()
				d.applyDefaultTax()
			}
		})
		items = append(items, widget.NewFormItem("", manualTaxCheck))
	}

	return widget.NewForm(items...)
}

func (d *ItemDialog) configureWidgets() {
	d.descEntry.SetPlaceHolder("Descripción del producto/servicio")
	d.qtyEntry.SetText("1")
	d.priceEntry.SetPlaceHolder("0.00")
	d.taxSelect.SetSelected("IVA 15%") // Default global

	d.applyDefaultTax()
}

func (d *ItemDialog) applyDefaultTax() {
	if d.defaultTaxRate >= 0 {
		var taxStr string
		switch d.defaultTaxRate {
		case 4:
			taxStr = "IVA 15%"
		case 10:
			taxStr = "IVA 13%"
		case 8:
			taxStr = "IVA 8%"
		case 5:
			taxStr = "IVA 5%"
		case 0:
			taxStr = "IVA 0%"
		case 6:
			taxStr = "No Objeto (6)"
		case 7:
			taxStr = "Exento (7)"
		}

		if taxStr != "" {
			d.taxSelect.SetSelected(taxStr)
			d.taxSelect.Disable()
		}
	}
}
