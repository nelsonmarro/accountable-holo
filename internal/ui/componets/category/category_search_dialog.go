package category

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
)

// CategorySearchDialog defines a reusable dialog for searching and selecting a category.
type CategorySearchDialog struct {
	mainWin            fyne.Window
	logger             *log.Logger
	catService         CategoryService
	onCategorySelected func(*domain.Category)

	// UI Components
	searchEntry  *widget.Entry
	categoryList *widget.List
	pagination   *componets.Pagination
	dl           dialog.Dialog

	// State
	categories []domain.Category
	totalCount int64 // Use int64 to match domain.PaginatedResult
	searchTerm string
	filterType *domain.CategoryType
}

// pageSizeOpts defines the available page size options for the user to select.
var pageSizeOpts = []string{"10", "20", "50"}

// NewCategorySearchDialog creates a new dialog for searching and selecting categories.
func NewCategorySearchDialog(
	mainWin fyne.Window,
	logger *log.Logger,
	catService CategoryService,
	onCategorySelected func(*domain.Category),
) *CategorySearchDialog {
	d := &CategorySearchDialog{
		mainWin:            mainWin,
		logger:             logger,
		catService:         catService,
		onCategorySelected: onCategorySelected,
	}
	d.searchEntry = widget.NewEntry()
	d.searchEntry.SetPlaceHolder("Buscar por nombre o tipo...")
	return d
}

func (d *CategorySearchDialog) SetFilterType(t domain.CategoryType) {
	d.filterType = &t
}

// Show creates and displays the dialog.
func (d *CategorySearchDialog) Show() {
	content := d.createContent()

	d.dl = dialog.NewCustom("Buscar Categoría", "Cerrar", content, d.mainWin)
	d.dl.Resize(fyne.NewSize(630, 550))

	// Initial load
	go d.loadCategories(1, d.pagination.GetPageSize())

	d.dl.Show()
}

// createContent builds the UI layout for the dialog.
func (d *CategorySearchDialog) createContent() fyne.CanvasObject {
	// --- Search Bar ---
	d.searchEntry.OnChanged = func(s string) {
		// Simple debounce
		time.AfterFunc(300*time.Millisecond, func() {
			if s == d.searchEntry.Text {
				d.filterCategories(s)
			}
		})
	}

	// --- Pagination ---
	d.pagination = componets.NewPagination(
		func() int {
			return int(d.totalCount)
		},
		d.loadCategories,
		pageSizeOpts...,
	)

	// --- List ---
	d.categoryList = widget.NewList(
		func() int {
			return len(d.categories)
		},
		d.makeListItem,
		d.updateListItem,
	)

	// --- Headers ---
	header := container.NewGridWithColumns(3,
		widget.NewLabel("Nombre"),
		widget.NewLabel("Tipo"),
		widget.NewLabel("Acción"),
	)

	// --- Layout ---
	topContainer := container.NewBorder(nil, nil, nil, nil, d.searchEntry)
	tableContainer := container.NewBorder(header, nil, nil, nil, d.categoryList)

	return container.NewBorder(
		container.NewVBox(topContainer, d.pagination),
		nil, nil, nil,
		tableContainer,
	)
}

// makeListItem creates the template UI for a single row in the list.
func (d *CategorySearchDialog) makeListItem() fyne.CanvasObject {
	selectBtn := widget.NewButtonWithIcon("Seleccionar", theme.ConfirmIcon(), nil)
	selectBtn.Importance = widget.HighImportance

	return container.NewGridWithColumns(3,
		widget.NewLabel("Template Name"),
		widget.NewLabel("Template Type"),
		selectBtn,
	)
}

// updateListItem populates a list row with the actual category data.
func (d *CategorySearchDialog) updateListItem(i widget.ListItemID, o fyne.CanvasObject) {
	if i >= len(d.categories) {
		return
	}
	cat := d.categories[i]
	row := o.(*fyne.Container)

	// Populate data
	nameLabel := row.Objects[0].(*widget.Label)
	nameLabel.SetText(cat.Name)

	typeLabel := row.Objects[1].(*widget.Label)
	typeLabel.SetText(string(cat.Type))

	// Set button action
	selectBtn := row.Objects[2].(*widget.Button)
	selectBtn.OnTapped = func() {
		d.logger.Printf("Category selected: %s (ID: %d)", cat.Name, cat.ID)
		if d.onCategorySelected != nil {
			d.onCategorySelected(&cat)
		}
		d.dl.Hide()
	}
}

// loadCategories fetches a paginated list of categories from the service.
func (d *CategorySearchDialog) loadCategories(page int, pageSize int) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := d.catService.GetSelectablePaginatedCategories(ctx, page, pageSize, d.searchTerm)
	if err != nil {
		d.logger.Printf("Error loading categories for search dialog: %v", err)
		// Show error on the main thread
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("no se pudieron cargar las categorías: %w", err), d.mainWin)
		})
		return
	}

	// Update state
	if d.filterType != nil {
		var filtered []domain.Category
		for _, cat := range result.Data {
			if cat.Type == *d.filterType {
				filtered = append(filtered, cat)
			}
		}
		d.categories = filtered
		d.totalCount = int64(len(filtered))
	} else {
		d.categories = result.Data
		d.totalCount = result.TotalCount
	}

	// Refresh UI on the main thread
	fyne.Do(func() {
		d.pagination.Refresh()
		d.categoryList.Refresh()
	})
}

// filterCategories is called when the search term changes.
func (d *CategorySearchDialog) filterCategories(filter string) {
	d.searchTerm = filter
	// Reset to page 1 for a new search
	d.loadCategories(1, d.pagination.GetPageSize())
}
