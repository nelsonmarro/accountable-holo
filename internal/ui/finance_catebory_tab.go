package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	data := [][]string{
		{"Nombre", "Tipo"},
		{"bottom left", "bottom right"},
	}

	ui.categoryTable = widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label, ok := o.(*widget.Label)
			if !ok {
				return
			}
			label.SetText(data[i.Row][i.Col])
		})

	// containers
	headerArea := container.NewVBox(
		container.NewCenter(title),
		container.NewHBox(layout.NewSpacer(), catAddBtn),
	)
	mainContent := container.NewBorder(container.NewPadded(headerArea), nil, nil, nil, ui.categoryTable)

	return container.NewScroll(mainContent)
}
