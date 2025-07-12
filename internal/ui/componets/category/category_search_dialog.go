package category

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
)

// CategorySearchDialog defines a reusable dialog for searching and selecting a category.
type CategorySearchDialog struct {
	mainWin            fyne.Window
	logger             *log.Logger
	catService         CategoryService
	onCategorySelected func(domain.Category)

	// UI Components
	searchEntry  *widget.Entry
	categoryList *widget.List
	pagination   *componets.Pagination
	dl           dialog.Dialog

	// State
	categories []domain.Category
	totalCount int
	searchTerm string
}
