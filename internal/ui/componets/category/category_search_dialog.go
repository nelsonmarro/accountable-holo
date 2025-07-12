package category

import (
	"context"
	"log"
	"time"

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

var pageSizeOpts = []string{"5", "10", "20", "50", "100"}

func NewCategorySearchDialog(
	mainWin fyne.Window,
	looger *log.Logger,
	catService CategoryService,
	onCategorySelected func(domain.Category),
) *CategorySearchDialog {
	d := &CategorySearchDialog{
		mainWin:            mainWin,
		logger:             looger,
		catService:         catService,
		onCategorySelected: onCategorySelected,
	}
	d.searchEntry = widget.NewEntry()
	d.searchEntry.SetPlaceHolder("Buscar...")

	return d
}

func (d *CategorySearchDialog) Show() {
	// Create the dialog's content
	content := d.createContent()

	// Create a new custom dialog
	d.dl = dialog.NewCustom("Buscar Categorias", "Cerrar", content, d.mainWin)
	d.dl.Resize(fyne.NewSize(450, 530))

	// Fetch the first page of categories
	go d.loadCategories(1, d.pagination.GetPageSize())

	d.dl.Show()
}

func (d *CategorySearchDialog) createContent() fyne.CanvasObject {
	// Search Bar
	d.searchEntry.OnChanged = func(filter string) { d.filterCategories(filter) }

	return widget.NewLabel("UI content goes here")
}

const CategoryPageSize = 10

func (d *CategorySearchDialog) loadCategories(page int, pageSize int) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := d.catService.GetPaginatedCategories(ctx, page, pageSize, d.searchTerm)
	if err != nil {
		dialog.ShowError(err, d.mainWin)
		return
	}

	d.categories = result.Data

	fyne.Do(func() {
		d.categoryList.Refresh()
		d.pagination.Refresh()
	})
}

func (d *CategorySearchDialog) filterCategories(filter string) {
	d.searchTerm = filter
	d.loadCategories(1, d.pagination.GetPageSize())
}
