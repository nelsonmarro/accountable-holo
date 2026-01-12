package transaction

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func TransactionForm(
	description fyne.CanvasObject,
	amount fyne.CanvasObject,
	date fyne.CanvasObject,
	category fyne.CanvasObject,
	attachment fyne.CanvasObject,
	isRecurring fyne.CanvasObject,
	interval fyne.CanvasObject,
	isRecurringLabel fyne.CanvasObject,
	taxSelect fyne.CanvasObject, // New
	taxPayerSelect fyne.CanvasObject, // New
	subtotalLabel fyne.CanvasObject, // New: Display calculated subtotal
	taxAmountLabel fyne.CanvasObject, // New: Display calculated tax
) []*widget.FormItem {
	addFormValidation(description, amount, category)

	return []*widget.FormItem{
		{Text: "Descripción", Widget: description},
		{Text: "Monto Total", Widget: amount}, // Clarify it's Total
		{Text: "Impuestos", Widget: taxSelect}, // New
		{Text: "Subtotal", Widget: subtotalLabel}, // New (Info)
		{Text: "IVA", Widget: taxAmountLabel}, // New (Info)
		{Text: "Cliente", Widget: taxPayerSelect}, // New
		{Text: "Fecha", Widget: date},
		{Text: "Categoria", Widget: category},
		{Text: "Adjunto", Widget: attachment},
		{Text: "", Widget: container.NewBorder(nil, nil, isRecurringLabel, nil, isRecurring)},
		{Text: "Frecuencia", Widget: interval},
	}
}

func addFormValidation(
	description fyne.CanvasObject,
	amount fyne.CanvasObject,
	category fyne.CanvasObject,
) {
	descEntry, ok := description.(*widget.Entry)
	if ok {
		descriptionValidator := uivalidators.NewValidator()
		descriptionValidator.Required()
		descriptionValidator.MinLength(3)
		descEntry.Validator = descriptionValidator.Validate
	}

	amtEntry, ok := amount.(*widget.Entry)
	if ok {
		amountValidator := uivalidators.NewValidator()
		amountValidator.Required()
		amountValidator.IsFloat()
		amtEntry.Validator = amountValidator.Validate
	}

	catSelect, ok := category.(*widget.SelectEntry)
	if ok {
		categoryValidator := uivalidators.NewValidator()
		categoryValidator.Required()
		catSelect.Validator = categoryValidator.Validate
	}
}
