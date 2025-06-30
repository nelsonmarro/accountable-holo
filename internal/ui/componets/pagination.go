// Package componets provides reusable widgets for the app.
package componets

import "fyne.io/fyne/v2/widget"

// Pagination is a custom widget for navigating through pages of data.
type Pagination struct {
	widget.BaseWidget

	TotalItems  int
	PageSize    int
	CurrentPage int

	// OnPageChanged is a callback function that is triggered when the page changes.
	OnPageChange func(page int)
}
