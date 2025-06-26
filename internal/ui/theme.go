package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AppTheme is our custom theme. It embeds a default Fyne theme and
// allows us to override specific parts, like font sizes.
type AppTheme struct {
	fyne.Theme
}

// NewAppTheme creates a new instance of our custom theme.
func NewAppTheme() fyne.Theme {
	// We can choose to wrap the default theme.
	return &AppTheme{Theme: theme.DefaultTheme()}
}

// Size overrides the Size method of the embedded theme.
func (t *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	// Use a switch to check which size is being requested.
	switch name {
	case theme.SizeNameText:
		// Make the standard text size slightly larger than the default.
		return 15
	case theme.SizeNameHeadingText:
		// Define a new size for headings that is larger and bold.
		// (The bold part is handled by the widget, like widget.RichText).
		return 24
	default:
		// For all other sizes, fall back to the embedded theme's default implementation.
		// This is CRITICAL to ensure the rest of the UI looks correct.
		return t.Theme.Size(name)
	}
}
