package account

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/application/uivalidators"
)

func AccountForm(
	nameEntry fyne.CanvasObject,
	tipoSelect fyne.CanvasObject,
	amountEntry fyne.CanvasObject,
	numberEntry fyne.CanvasObject,
) []*widget.FormItem {
	addFormValidation(nameEntry, amountEntry, tipoSelect, numberEntry)

	return []*widget.FormItem{
		{Text: "Nombre", Widget: nameEntry},
		{Text: "Número", Widget: numberEntry},
		{Text: "Tipo", Widget: tipoSelect},
		{Text: "Monto Inicial", Widget: amountEntry},
	}
}

func addFormValidation(
	nameEntry fyne.CanvasObject,
	amountEntry fyne.CanvasObject,
	tipoSelect fyne.CanvasObject,
	numberEntry fyne.CanvasObject,
) {
	name, ok := nameEntry.(*widget.Entry)
	if ok {
		nameValidator := uivalidators.NewValidator()
		nameValidator.Required()
		nameValidator.MinLength(3)
		name.Validator = nameValidator.Validate
	}

	amount, ok := amountEntry.(*widget.Entry)
	if ok {
		amountValidator := uivalidators.NewValidator()
		amountValidator.Required()
		amountValidator.IsFloat()
		amount.Validator = amountValidator.Validate
	}

	tipo, ok := tipoSelect.(*widget.SelectEntry)
	if ok {
		tipoValidator := uivalidators.NewValidator()
		tipoValidator.Required()
		tipo.Validator = tipoValidator.Validate
	}

	number, ok := numberEntry.(*widget.Entry)
	if ok {
		numberValidator := uivalidators.NewValidator()
		numberValidator.Required()
		numberValidator.IsInt()
		number.Validator = numberValidator.Validate
	}
}
