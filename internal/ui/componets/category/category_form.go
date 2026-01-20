package category

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/application/uivalidators"
)

func CategoryForm(
	nameEntry fyne.CanvasObject,
	tipoSelect fyne.CanvasObject,
	monthlyBudgetEntry fyne.CanvasObject,
) []*widget.FormItem {
	addFormValidation(nameEntry, tipoSelect, monthlyBudgetEntry)

	return []*widget.FormItem{
		{Text: "Nombre", Widget: nameEntry},
		{Text: "Tipo", Widget: tipoSelect},
		{Text: "Presupuesto Mensual (Opcional)", Widget: monthlyBudgetEntry},
	}
}

func addFormValidation(
	nameEntry fyne.CanvasObject,
	tipoSelect fyne.CanvasObject,
	monthlyBudgetEntry fyne.CanvasObject,
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

	budget, ok := monthlyBudgetEntry.(*widget.Entry)
	if ok {
		budgetValidator := uivalidators.NewValidator()
		budgetValidator.IsFloat() // This likely enforces it IS a float. We need "If not empty, must be float"

		budget.Validator = func(s string) error {
			if s == "" {
				return nil
			}
			return budgetValidator.Validate(s)
		}
	}
}
