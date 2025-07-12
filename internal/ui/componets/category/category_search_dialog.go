package category

import (
	"context"
	"errors"
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
	d.fetchCategories(1, "")

	d.dl.Show()
}

func (d *CategorySearchDialog) createContent() fyne.CanvasObject {
	// Create and arrange UI Components

	return widget.NewLabel("UI content goes here")
}

const CategoryPageSize = 10

func (d *CategorySearchDialog) fetchCategories(page int, s string) {
	progressDialog := dialog.NewCustomWithoutButtons("Espere", widget.NewProgressBarInfinite(), d.mainWin)
	progressDialog.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		paginatedRes, err := d.catService.GetPaginatedCategories(ctx, page, CategoryPageSize, s)
		if err != nil {
			d.logger.Println("Error fetching categories:", err)
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(errors.New("error al cargar categorias"), d.mainWin)
			})
			return
		}

		d.categories = paginatedRes.Data
		d.totalCount = int(paginatedRes.TotalCount)
		d.searchTerm = s

		fyne.Do(func() {
			progressDialog.Hide()
			d.categoryList.Refresh()
			d.pagination.Refresh()
		})
	}()
}
