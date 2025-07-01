package account

import (
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

func AccountForm(
	nameEntry *widget.Entry,
	tipoSelect *widget.SelectEntry,
	amountEntry *widget.Entry,
	numberEntry *widget.Entry,
) []*widget.FormItem {
	tipoSelect.SetText(string(domain.SavingAccount))
	addFormValidation(nameEntry, amountEntry, tipoSelect, numberEntry)

	return []*widget.FormItem{
		{Text: "Nombre", Widget: nameEntry},
		{Text: "Número", Widget: numberEntry},
		{Text: "Tipo", Widget: tipoSelect},
		{Text: "Monto Inicial", Widget: amountEntry},
	}
}

func addFormValidation(
	nameEntry *widget.Entry,
	amountEntry *widget.Entry,
	tipoSelect *widget.SelectEntry,
	numberEntry *widget.Entry,
) {
	nameValidator := uivalidators.NewValidator()
	nameValidator.Required()
	nameValidator.MinLength(3)

	nameEntry.Validator = nameValidator.Validate

	amountValidator := uivalidators.NewValidator()
	amountValidator.Required()
	amountValidator.IsFloat()

	amountEntry.Validator = amountValidator.Validate

	tipoValidator := uivalidators.NewValidator()
	tipoValidator.Required()

	tipoSelect.Validator = tipoValidator.Validate

	numberValidator := uivalidators.NewValidator()
	numberValidator.Required()
	numberValidator.IsInt()

	numberEntry.Validator = numberValidator.Validate
}
