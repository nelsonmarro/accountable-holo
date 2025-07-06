package transaction

import (
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func TransactionForm(
	descriptionEntry *widget.Entry,
	amountEntry *widget.Entry,
	dateEntry *widget.Entry,
	categorySelect *widget.SelectEntry,
) []*widget.FormItem {
	dateEntry.SetPlaceHolder("YYYY-MM-DD")

	addFormValidation(descriptionEntry, amountEntry, dateEntry, categorySelect)
	return []*widget.FormItem{
		{Text: "Description", Widget: descriptionEntry},
		{Text: "Amount", Widget: amountEntry},
		{Text: "Date", Widget: dateEntry},
		{Text: "Category", Widget: categorySelect},
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
