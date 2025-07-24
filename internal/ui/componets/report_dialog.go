package componets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type ReportDialog struct {
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

func NewReportDialog(parentWindow fyne.Window, allCategories []domain.Category, applyCallback func(filters domain.TransactionFilters)) *ReportDialog {
	rd := &ReportDialog{
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

	// --- Build the form ---

	return rd
}
