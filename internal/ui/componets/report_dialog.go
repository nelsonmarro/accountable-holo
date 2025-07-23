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
