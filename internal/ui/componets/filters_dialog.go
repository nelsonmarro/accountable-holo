package componets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type FiltersDialog struct {
	parentWindow     fyne.Window
	startDateEntry   *widget.DateEntry
	endDateEntry     *widget.DateEntry
	categorySelect   *widget.SelectEntry
	typeSelect       *widget.SelectEntry
	descriptionEntry *widget.Entry
	allCategories    []domain.Category
	applyCallback    func(filters domain.TransactionFilters)
	dialog           dialog.Dialog
}

func NewFiltersDialog(parentWindow fyne.Window, allCategories []domain.Category, applyCallback func(filters domain.TransactionFilters)) *FiltersDialog {
	rd := &FiltersDialog{
		parentWindow:  parentWindow,
		allCategories: allCategories,
		applyCallback: applyCallback,
	}

	// Instantiate the dialog components
	rd.startDateEntry = widget.NewDateEntry()
	rd.endDateEntry = widget.NewDateEntry()
	rd.descriptionEntry = widget.NewEntry()
	rd.descriptionEntry.SetPlaceHolder("Filter by description")

	// Setup category selection
	categoryNames := []string{"All"}
	for _, cat := range allCategories {
		categoryNames = append(categoryNames, cat.Name)
	}

	rd.categorySelect = widget.NewSelectEntry(categoryNames)
	rd.categorySelect.SetText("All")

	// Setup type selection
	rd.typeSelect = widget.NewSelectEntry([]string{"All", string(domain.Income), string(domain.Outcome)})
	rd.typeSelect.SetText("All") // Default to "All"

	// --- Create form items ---
	formItems := []*widget.FormItem{
		{Text: "Fecha de Inicio", Widget: rd.startDateEntry},
		{Text: "Fecha de Fin", Widget: rd.endDateEntry},
		{Text: "Categoria", Widget: rd.categorySelect},
		{Text: "Tipo de Transacción", Widget: rd.typeSelect},
		{Text: "Descripción", Widget: rd.descriptionEntry},
	}

	// Create the formDialog
	callback := func(confirmed bool) {
		if !confirmed {
			return
		}

		filters := rd.buildFilters()
		rd.applyCallback(filters)
	}

	rd.dialog = dialog.NewForm(
		"Filtros Avanzados",
		"Aplicar",
		"Cancelar",
		formItems,
		callback,
		rd.parentWindow,
	)

	return rd
}

func (rd *FiltersDialog) Show() {
	rd.dialog.Resize(fyne.NewSize(480, 380))
	rd.dialog.Show()
}

func (rd *FiltersDialog) buildFilters() domain.TransactionFilters {
	filters := domain.TransactionFilters{}

	if !rd.startDateEntry.Date.IsZero() {
		filters.StartDate = rd.startDateEntry.Date
	}
	if !rd.endDateEntry.Date.IsZero() {
		filters.EndDate = rd.endDateEntry.Date
	}

	if rd.descriptionEntry.Text != "" {
		filters.Description = &rd.descriptionEntry.Text
	}

	selectedType := rd.typeSelect.Text
	if selectedType == string(domain.Income) || selectedType == string(domain.Outcome) {
		ct := domain.CategoryType(selectedType)
		filters.CategoryType = &ct
	}

	selectedCatName := rd.categorySelect.Text
	if selectedCatName != "All" && selectedCatName != "" {
		for _, cat := range rd.allCategories {
			if cat.Name == selectedCatName {
				filters.CategoryID = &cat.ID
				break
			}
		}
	}

	return filters
}
