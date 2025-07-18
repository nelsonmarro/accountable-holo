// Package componets provides reusable widgets for the app.
package componets

import (
	"fmt"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Pagination is a custom widget for navigating through pages of data.
type Pagination struct {
	widget.BaseWidget

	currentPage     int
	pageSizeOptions []string
	pageSize        int

	// onPageChanged is a callback function that is triggered when the page changes.
	onPageChanged func(page, pageSize int)
	getTotalCount func() (totalCount int)
}

// NewPagination creates a new pagination widget.
// onPageChanged will be called with the new page number when the user navigates.
func NewPagination(
	getTotalCount func() (totalCount int),
	onPageChanged func(page, pageSize int),
	pageSizeOptions ...string,
) *Pagination {
	if pageSizeOptions == nil {
		pageSizeOptions = []string{"5", "10", "20", "50", "100"}
	}
	initPageSize, _ := strconv.Atoi(pageSizeOptions[0])
	p := &Pagination{
		getTotalCount:   getTotalCount,
		currentPage:     1,
		onPageChanged:   onPageChanged,
		pageSizeOptions: pageSizeOptions,
		pageSize:        initPageSize,
	}

	p.ExtendBaseWidget(p)
	return p
}

func (p *Pagination) GetPageSize() int {
	if p.pageSize <= 0 {
		return 10 // Default page size if not set
	}
	return p.pageSize
}

// CreateRenderer is the entry point for Fyne to create the visual component.
func (p *Pagination) CreateRenderer() fyne.WidgetRenderer {
	p.ExtendBaseWidget(p)
	// A reference to the widget is passed to the renderer.
	r := &paginationRenderer{widget: p}

	// Create all the UI components one time.
	r.firstBtn = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), r.onFirst)
	r.prevBtn = widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), r.onPrev)
	r.nextBtn = widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), r.onNext)
	r.lastBtn = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), r.onLast)
	r.pageSizeSelect = widget.NewSelect(p.pageSizeOptions, r.onSelectPageSize)
	r.pageSizeSelect.Selected = p.pageSizeOptions[0] // Default to the first option

	// Create the 5 buttons for the page numbers. We'll set their text later.
	for i := range 5 {
		idx := i
		btn := widget.NewButton(fmt.Sprintf("%d", idx+1), func() {
			r.onPageTapped(idx + 1)
		})
		r.pageBtns = append(r.pageBtns, btn)
	}

	// Create the main layout container.
	r.layout = container.NewHBox(
		r.firstBtn, r.prevBtn,
		layout.NewSpacer(),
		r.pageBtns[0], r.pageBtns[1], r.pageBtns[2], r.pageBtns[3], r.pageBtns[4],
		layout.NewSpacer(),
		r.nextBtn, r.lastBtn,
		r.pageSizeSelect,
		widget.NewLabel("Items por página"),
	)

	// Initial refresh to set the correct state.
	r.Refresh()
	return r
}

// paginationRenderer is the renderer for the Pagination widget.
type paginationRenderer struct {
	widget         *Pagination
	firstBtn       *widget.Button
	prevBtn        *widget.Button
	nextBtn        *widget.Button
	lastBtn        *widget.Button
	pageSizeSelect *widget.Select
	pageBtns       []*widget.Button
	layout         *fyne.Container
}

// Layout Tells Fyne how to size and position the objetcs in a widget.
func (r *paginationRenderer) Layout(size fyne.Size) {
	r.layout.Resize(size)
}

// MinSize calculates the minimun size required for the widget.
func (r *paginationRenderer) MinSize() fyne.Size {
	return r.layout.MinSize()
}

// Objects returns all the canvas objects that make up the widget.
func (r *paginationRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.layout}
}

// Destroy is called when the widget is no longer needed.
func (r *paginationRenderer) Destroy() {}

// -- Navigation Handlers --

func (r *paginationRenderer) onFirst() {
	r.navigateTo(1)
}

func (r *paginationRenderer) onPrev() {
	r.navigateTo(r.widget.currentPage - 1)
}

func (r *paginationRenderer) onNext() {
	r.navigateTo(r.widget.currentPage + 1)
}

func (r *paginationRenderer) onLast() {
	r.navigateTo(r.totalPages())
}

func (r *paginationRenderer) onSelectPageSize(size string) {
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		fmt.Println("Error parsing page size:", err)
		return
	}
	r.widget.pageSize = sizeInt
	r.navigateTo(1)
}

func (r *paginationRenderer) onPageTapped(page int) {
	r.navigateTo(page)
}

func (r *paginationRenderer) navigateTo(page int) {
	if r.widget.onPageChanged != nil {
		r.widget.currentPage = page
		r.widget.onPageChanged(page, r.widget.pageSize) // Notify the main app
		r.Refresh()                                     // Update the pagination widget itself
	}
}

// -- Refresh Logic --
func (r *paginationRenderer) Refresh() {
	totalPages := r.totalPages()

	// Determine the 'sliding window' of page numbers to show.
	startPage, endPage := r.calculatePageRange(totalPages)

	// Update the page number buttons.
	pageNumber := startPage
	for i := range 5 {
		btn := r.pageBtns[i]
		if pageNumber <= endPage && pageNumber > 0 {
			page := pageNumber // Capture the current page number for closure
			btn.SetText(fmt.Sprintf("%d", pageNumber))
			btn.OnTapped = func() {
				r.onPageTapped(page)
			}
			if pageNumber == r.widget.currentPage {
				btn.Importance = widget.HighImportance // Highlight the current page
			} else {
				btn.Importance = widget.MediumImportance
			}
			btn.Show()
		} else {
			// Hide buttons if there are fewer than 5 pages to show
			btn.Hide()
		}
		pageNumber++
	}

	// Update the state of navigation buttons (first, prev, next, last).
	r.firstBtn.Disable()
	r.prevBtn.Disable()
	if r.widget.currentPage > 1 {
		r.firstBtn.Enable()
		r.prevBtn.Enable()
	}

	r.lastBtn.Disable()
	r.nextBtn.Disable()
	if r.widget.currentPage < totalPages {
		r.lastBtn.Enable()
		r.nextBtn.Enable()
	}
}

func (r *paginationRenderer) totalPages() int {
	totalCount := r.widget.getTotalCount()
	if totalCount == 0 || r.widget.pageSize == 0 {
		return 1
	}

	return int(math.Ceil(float64(totalCount) / float64(r.widget.pageSize)))
}

func (r *paginationRenderer) calculatePageRange(totalPages int) (int, int) {
	start := r.widget.currentPage - 2
	end := r.widget.currentPage + 2

	if start < 1 {
		diff := 1 - start
		start += diff
		end += diff
	}

	if end > totalPages {
		start -= (end - totalPages)
		end = totalPages
	}

	return start, end
}
