package transaction

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func TransactionForm(
	description fyne.CanvasObject,
	amount fyne.CanvasObject,
	date fyne.CanvasObject,
	category fyne.CanvasObject,
) []*widget.FormItem {
	dtEntry, ok := date.(*widget.Entry)
	if ok {
		dtEntry.SetPlaceHolder("YYYY-MM-DD")
	}

	addFormValidation(description, amount, date, category)

	return []*widget.FormItem{
		{Text: "Descripción", Widget: description},
		{Text: "Monto", Widget: amount},
		{Text: "Fecha", Widget: date},
		{Text: "Categoria", Widget: category},
	}
}

func addFormValidation(
	description fyne.CanvasObject,
	amount fyne.CanvasObject,
	date fyne.CanvasObject,
	category fyne.CanvasObject,
) {
	descEntry, ok := description.(*widget.Entry)

	if !ok {
		descriptionValidator := uivalidators.NewValidator()
		descriptionValidator.Required()
		descriptionValidator.MinLength(3)
		descEntry.Validator = descriptionValidator.Validate
	}

	amtEntry := amount.(*widget.Entry)
	dtEntry := date.(*widget.Entry)
	catSelect := category.(*widget.SelectEntry)

	amountValidator := uivalidators.NewValidator()
	amountValidator.Required()
	amountValidator.IsFloat()
	amount.Validator = amountValidator.Validate

	dateValidator := uivalidators.NewValidator()
	dateValidator.Required()
	date.Validator = dateValidator.Validate

	categoryValidator := uivalidators.NewValidator()
	categoryValidator.Required()
	category.Validator = categoryValidator.Validate
}
