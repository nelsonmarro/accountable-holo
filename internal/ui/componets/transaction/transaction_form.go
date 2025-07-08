package transaction

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func TransactionForm(
	descriptionEntry fyne.CanvasObject,
	amountEntry fyne.CanvasObject,
	dateEntry fyne.CanvasObject,
	categorySelect fyne.CanvasObject,
) []*widget.FormItem {
	// Assert canvas objects to their concrete widget types
	descEntry := descriptionEntry.(*widget.Entry)
	amtEntry := amountEntry.(*widget.Entry)
	dtEntry := dateEntry.(*widget.Entry)
	catSelect := categorySelect.(*widget.SelectEntry)

	dtEntry.SetPlaceHolder("YYYY-MM-DD")

	// Pass the asserted widgets to the validation function
	addFormValidation(descEntry, amtEntry, dtEntry, catSelect)

	return []*widget.FormItem{
		{Text: "Descripción", Widget: descEntry},
		{Text: "Monto", Widget: amtEntry},
		{Text: "Fecha", Widget: dtEntry},
		{Text: "Categoria", Widget: catSelect},
	}
}

func addFormValidation(
	descriptionEntry *widget.Entry,
	amountEntry *widget.Entry,
	dateEntry *widget.Entry,
	categorySelect *widget.SelectEntry,
) {
	descriptionValidator := uivalidators.NewValidator()
	descriptionValidator.Required()
	descriptionValidator.MinLength(3)
	descriptionEntry.Validator = descriptionValidator.Validate

	amountValidator := uivalidators.NewValidator()
	amountValidator.Required()
	amountValidator.IsFloat()
	amountEntry.Validator = amountValidator.Validate

	dateValidator := uivalidators.NewValidator()
	dateValidator.Required()
	dateEntry.Validator = dateValidator.Validate

	categoryValidator := uivalidators.NewValidator()
	categoryValidator.Required()
	categorySelect.Validator = categoryValidator.Validate
}
