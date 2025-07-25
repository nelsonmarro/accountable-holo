package componets

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/category"
)

type FiltersDialog struct {
	catService        CategoryService
	parentWindow      fyne.Window
	startDateEntry    *widget.DateEntry
	endDateEntry      *widget.DateEntry
	categoryLabel     *widget.Label
	searchCategoryBtn *widget.Button
	selectedCategory  *domain.Category
	typeSelect        *widget.SelectEntry
	descriptionEntry  *widget.Entry
	applyCallback     func(filters domain.TransactionFilters)
	dialog            dialog.Dialog
	logger            *log.Logger
}

func NewFiltersDialog(
	parentWindow fyne.Window,
	catService CategoryService,
	applyCallback func(filters domain.TransactionFilters),
	logger *log.Logger,
) *FiltersDialog {
	rd := &FiltersDialog{
		parentWindow:  parentWindow,
		catService:    catService,
		applyCallback: applyCallback,
		logger:        logger,
	}
	// Instantiate the dialog components

	rd.categoryLabel = widget.NewLabel("Categoría")
	rd.startDateEntry = widget.NewDateEntry()
	rd.endDateEntry = widget.NewDateEntry()
	rd.descriptionEntry = widget.NewEntry()
	rd.descriptionEntry.SetPlaceHolder("Filter by description")

	rd.searchCategoryBtn = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchDialog := category.NewCategorySearchDialog(
			rd.parentWindow,
			rd.logger,
			rd.catService,
			func(cat *domain.Category) {
				rd.selectedCategory = cat
				rd.categoryLabel.SetText(cat.Name)
			},
		)
		searchDialog.Show()
	})
	categoryContainer := container.NewBorder(nil, nil, nil, rd.searchCategoryBtn, rd.categoryLabel)

	// Setup type selection
	rd.typeSelect = widget.NewSelectEntry([]string{"All", string(domain.Income), string(domain.Outcome)})
	rd.typeSelect.SetText("All") // Default to "All"

	// --- Create form items ---
	formItems := []*widget.FormItem{
		{Text: "Fecha de Inicio", Widget: rd.startDateEntry},
		{Text: "Fecha de Fin", Widget: rd.endDateEntry},
		{Text: "Categoria", Widget: categoryContainer},
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
