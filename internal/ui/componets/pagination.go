// Package componets provides reusable widgets for the app.
package componets

import (
	"fyne.io/fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Pagination is a custom widget for navigating through pages of data.
type Pagination struct {
	widget.BaseWidget

	TotalItems  int
	PageSize    int
	CurrentPage int

	// OnPageChanged is a callback function that is triggered when the page changes.
	OnPageChange func(page int)
}

// NewPagination creates a new pagination widget.
// onPageChanged will be called with the new page number when the user navigates.
func NewPagination(totalItems, pageSize int, onPageChanged func(page int)) *Pagination {
	p := &Pagination{
		TotalItems:   totalItems,
		PageSize:     pageSize,
		CurrentPage:  1,
		OnPageChange: onPageChanged,
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
	}
}
