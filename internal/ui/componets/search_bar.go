// Package componets provides reusable widgets for the app.
package componets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SearchBar is a custom widget for filtering data.
type SearchBar struct {
	widget.BaseWidget
	entry     *widget.Entry
	OnChanged func(string)
}

// NewSearchBar creates a new search bar widget.
func NewSearchBar(onChanged func(string)) *SearchBar {
	s := &SearchBar{
		OnChanged: onChanged,
	}
	s.entry = widget.NewEntry()
	s.ExtendBaseWidget(s)
	return s
}

// CreateRenderer is the entry point for Fyne to create the visual component.
func (s *SearchBar) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)

	s.entry.SetPlaceHolder("Search...")
	s.entry.OnChanged = s.OnChanged

	icon := widget.NewIcon(theme.SearchIcon())

	// Use a border layout to place the icon on the left of the entry
	content := container.NewBorder(nil, nil, icon, nil, s.entry)

	r := &searchBarRenderer{
		widget:  s,
		entry:   s.entry,
		icon:    icon,
		content: content,
	}

	return r
}

func (s *SearchBar) SetText(text string) {
	s.entry.SetText(text)
}

// searchBarRenderer is the renderer for the SearchBar widget.
type searchBarRenderer struct {
	widget  *SearchBar
	entry   *widget.Entry
	icon    *widget.Icon
	content *fyne.Container
}

// Layout tells Fyne how to size and position the objects in a widget.
func (r *searchBarRenderer) Layout(size fyne.Size) {
	r.content.Resize(size)
}

// MinSize calculates the minimum size required for the widget.
func (r *searchBarRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

// Objects returns all the canvas objects that make up the widget.
func (r *searchBarRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

// Destroy is called when the widget is no longer needed.
func (r *searchBarRenderer) Destroy() {}

// Refresh is called when the widget needs to be redrawn.
func (r *searchBarRenderer) Refresh() {
	r.content.Refresh()
}
