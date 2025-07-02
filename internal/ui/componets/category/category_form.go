package category

import (
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func CategoryForm(
	nameEntry *widget.Entry,
	tipoSelect *widget.SelectEntry,
) []*widget.FormItem {
	addFormValidation(nameEntry, tipoSelect)

	return []*widget.FormItem{
		{Text: "Nombre", Widget: nameEntry},
		{Text: "Tipo", Widget: tipoSelect},
	}
}

func addFormValidation(
	nameEntry *widget.Entry,
	tipoSelect *widget.SelectEntry,
) {
	nameValidator := uivalidators.NewValidator()
	nameValidator.Required()
	nameValidator.MinLength(3)
	nameEntry.Validator = nameValidator.Validate

	tipoValidator := uivalidators.NewValidator()
	tipoValidator.Required()
	tipoSelect.Validator = tipoValidator.Validate
}
