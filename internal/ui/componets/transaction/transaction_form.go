package transaction

import (
	"time"

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

	dtEntry, ok := date.(*widget.Entry)
	if ok {
		dateValidator := uivalidators.NewValidator()
		dateValidator.Required()
		dateValidator.IsDate()
		dateValidator.MaxDate(time.Now())
		dtEntry.Validator = dateValidator.Validate
	}

	catSelect, ok := category.(*widget.SelectEntry)
	if ok {
		categoryValidator := uivalidators.NewValidator()
		categoryValidator.Required()
		catSelect.Validator = categoryValidator.Validate
	}
}
