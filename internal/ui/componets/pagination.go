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

	TotalItems  int
	PageSize    int
	CurrentPage int

	// OnPageChanged is a callback function that is triggered when the page changes.
	OnPageChanged func(page int)
}

// NewPagination creates a new pagination widget.
// onPageChanged will be called with the new page number when the user navigates.
func NewPagination(totalItems, pageSize int, onPageChanged func(page int)) *Pagination {
	p := &Pagination{
		TotalItems:    totalItems,
		PageSize:      pageSize,
		CurrentPage:   1,
		OnPageChanged: onPageChanged,
	}

	p.ExtendBaseWidget(p)
	return p
}

// CreateRenderer is the entry point for Fyne to create the visual component.
func (p *Pagination) CreateRenderer() fyne.WidgetRenderer {
	// A reference to the widget is passed to the renderer.
	r := &paginationRenderer{widget: p}

	// Create all the UI components one time.
	r.firstBtn = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), r.onFirst)
	r.prevBtn = widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), r.onPrev)
	r.nextBtn = widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), r.onNext)
	r.lastBtn = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), r.onLast)

	// Create the 5 buttons for the page numbers. We'll set their text later.
	for i := 0; i < 5; i++ {
		// The tap handler is set here. It captures the index `i` to know which button was pressed.
		// We use this to calculate the page number later.
		idx := i
		btn := widget.NewButton(fmt.Sprintf("%s", idx+1), func() {
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
	)

	// Initial refresh to set the correct state.
	r.Refresh()

	return r
}

// paginationRenderer is the renderer for the Pagination widget.
type paginationRenderer struct {
	widget   *Pagination
	firstBtn *widget.Button
	prevBtn  *widget.Button
	nextBtn  *widget.Button
	lastBtn  *widget.Button
	pageBtns []*widget.Button
	layout   *fyne.Container
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
	r.navigateTo(r.widget.CurrentPage - 1)
}

func (r *paginationRenderer) onNext() {
	r.navigateTo(r.widget.CurrentPage + 1)
}

func (r *paginationRenderer) onLast() {
	r.navigateTo(r.totalPages())
}

func (r *paginationRenderer) onPageTapped(btnIndex int) {
	page, _ := strconv.Atoi(r.pageBtns[btnIndex].Text)
	r.navigateTo(page)
}

func (r *paginationRenderer) navigateTo(page int) {
	if r.widget.OnPageChanged != nil {
		r.widget.CurrentPage = page
		r.widget.OnPageChanged(page) // Notify the main app
		r.Refresh()                  // Update the pagination widget itself
	}
}

// -- Refresh Logic --

// Refresh is the most important method. It's called to update the UI
// based on the widget's current state.
func (r *paginationRenderer) Refresh() {
	totalPages := r.totalPages()

	// Determine the 'sliding window' of page numbers to show.
	startPage, endPage := r.calculatePageRange(totalPages)

	// Update the page number buttons.
	pageNumber := startPage
	for i := 0; i < 5; i++ {
		btn := r.pageBtns[i]
		if pageNumber <= endPage {
			btn.SetText(fmt.Sprintf("%d", pageNumber))
			if pageNumber == r.widget.CurrentPage {
				btn.Importance = widget.HighImportance // Highlight the current page
			} else {
				btn.Importance = widget.MediumImportance
			}
			btn.Show()
			pageNumber++
		} else {
			// Hide buttons if there are fewer than 5 pages to show
			btn.Hide()
		}
	}

	// Update the state of navigation buttons (first, prev, next, last).
	r.firstBtn.Disable()
	r.prevBtn.Disable()
	if r.widget.CurrentPage > 1 {
		r.firstBtn.Enable()
		r.prevBtn.Enable()
	}

	r.lastBtn.Disable()
	r.nextBtn.Disable()
	if r.widget.CurrentPage < totalPages {
		r.lastBtn.Enable()
		r.nextBtn.Enable()
	}
}

func (r *paginationRenderer) totalPages() int {
	if r.widget.TotalItems == 0 || r.widget.PageSize == 0 {
		return 1
	}

	return int(math.Ceil(float64(r.widget.TotalItems) / float64(r.widget.PageSize)))
}
