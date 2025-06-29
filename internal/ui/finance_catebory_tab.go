package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeCategoryUI() fyne.CanvasObject {
	// UI widgets
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Categorias",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText, // Use the heading size from our custom theme
			Alignment: fyne.TextAlignCenter,
		},
	})

	catAddBtn := widget.NewButtonWithIcon("Agregar Categoría", theme.ContentAddIcon(), func() {})
	catAddBtn.Importance = widget.HighImportance

	// containers
}
