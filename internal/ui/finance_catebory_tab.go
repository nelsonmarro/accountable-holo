package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
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

	categories := []domain.Category{
		{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Sales Revenue", Type: "income"},
		{BaseEntity: domain.BaseEntity{ID: 2}, Name: "Office Supplies", Type: "outcome"},
		{BaseEntity: domain.BaseEntity{ID: 3}, Name: "Consulting Services", Type: "income"},
		{BaseEntity: domain.BaseEntity{ID: 4}, Name: "Software Subscriptions", Type: "outcome"},
	}

	// containers
	headerArea := container.NewVBox(
		container.NewCenter(title),
		container.NewHBox(layout.NewSpacer(), catAddBtn),
	)
	mainContent := container.NewBorder(container.NewPadded(headerArea), nil, nil, nil, ui.categoryList)

	return container.NewScroll(mainContent)
}
