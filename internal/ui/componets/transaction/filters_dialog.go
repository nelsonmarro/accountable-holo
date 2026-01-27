package transaction

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
	"github.com/nelsonmarro/verith/internal/ui/componets/category"
)

type FiltersDialog struct {
	catService        CategoryService
	parentWindow      fyne.Window
	startDateEntry    *componets.LatinDateEntry
	endDateEntry      *componets.LatinDateEntry
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
	return &FiltersDialog{
		parentWindow:  parentWindow,
		catService:    catService,
		applyCallback: applyCallback,
		logger:        logger,
	}
}

func (rd *FiltersDialog) Show() {
	rd.startDateEntry = componets.NewLatinDateEntry(rd.parentWindow)
	// rd.startDateEntry.SetDate(nil) // LatinDateEntry no soporta nil directo en SetDate por ahora, pero Date empieza nil

	rd.endDateEntry = componets.NewLatinDateEntry(rd.parentWindow)
	// rd.endDateEntry.SetDate(nil)

	rd.descriptionEntry = widget.NewEntry()
	rd.descriptionEntry.SetPlaceHolder("Filter by description")

	rd.categoryLabel = widget.NewLabel("Categoría")

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
	form := widget.NewForm(formItems...)

	// Create the formDialog
	callback := func(confirmed bool) {
		if !confirmed {
			return
		}

		filters := rd.buildFilters()
		rd.applyCallback(filters)
	}

	rd.dialog = dialog.NewCustomConfirm(
		"Filtros Avanzados",
		"Aplicar",
		"Cancelar",
		form,
		callback,
		rd.parentWindow,
	)

	rd.dialog.Resize(fyne.NewSize(480, 380))
	rd.dialog.Show()
}

func (rd *FiltersDialog) buildFilters() domain.TransactionFilters {
	filters := domain.TransactionFilters{}

	if rd.startDateEntry.Date != nil {
		filters.StartDate = rd.startDateEntry.Date
	}
	if rd.endDateEntry.Date != nil {
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

	if rd.selectedCategory != nil {
		filters.CategoryID = &rd.selectedCategory.ID
	}

	return filters
}
