package category

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func CategoryForm(
	nameEntry fyne.CanvasObject,
	tipoSelect fyne.CanvasObject,
) []*widget.FormItem {
	addFormValidation(nameEntry, tipoSelect)

	return []*widget.FormItem{
		{Text: "Nombre", Widget: nameEntry},
		{Text: "Tipo", Widget: tipoSelect},
	}
}

func addFormValidation(
	nameEntry fyne.CanvasObject,
	tipoSelect fyne.CanvasObject,
) {
	name, ok := nameEntry.(*widget.Entry)
	if ok {
		nameValidator := uivalidators.NewValidator()
		nameValidator.Required()
		nameValidator.MinLength(3)
		name.Validator = nameValidator.Validate
	}

	tipo, ok := tipoSelect.(*widget.SelectEntry)
	if ok {
		tipoValidator := uivalidators.NewValidator()
		tipoValidator.Required()
		tipo.Validator = tipoValidator.Validate
	}
}
