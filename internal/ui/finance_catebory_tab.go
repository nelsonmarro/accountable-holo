package ui

import (
	"fmt"
	"log"

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

	header := container.NewGridWithColumns(4,
		widget.NewLabel("ID"),
		widget.NewLabel("Name"),
		widget.NewLabel("Type"),
		widget.NewLabel("Actions"),
	)

	list := widget.NewList(
		func() int {
			return len(categories)
		},
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(4,
				widget.NewLabel("template id"),
				widget.NewLabel("template name"),
				widget.NewLabel("template type"),
				container.NewHBox( // A container for our action buttons
					widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil),
					widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			category := categories[i]

			rowContainer := o.(*fyne.Container)

			idLabel := rowContainer.Objects[0].(*widget.Label)
			idLabel.SetText(fmt.Sprintf("%d", category.ID))

			nameLabel := rowContainer.Objects[1].(*widget.Label)
			nameLabel.SetText(category.Name)

			typeLabel := rowContainer.Objects[2].(*widget.Label)
			typeLabel.SetText(string(category.Type))

			actionsContainer := rowContainer.Objects[3].(*fyne.Container)
			editBtn := actionsContainer.Objects[0].(*widget.Button)
			deleteBtn := actionsContainer.Objects[1].(*widget.Button)

			editBtn.OnTapped = func() {
				log.Printf("Edit button tapped for category ID: %d, Name: %s", category.ID, category.Name)
			}
			deleteBtn.OnTapped = func() {
				log.Printf("Delete button tapped for category ID: %d, Name: %s", category.ID, category.Name)
			}
		},
	)

	// containers
	headerArea := container.NewVBox(
		container.NewCenter(title),
		container.NewHBox(layout.NewSpacer(), catAddBtn),
	)
	tableContainer := container.NewBorder(header, nil, nil, nil, ui.categoryList)
	mainContent := container.NewBorder(container.NewPadded(headerArea), nil, nil, nil, tableContainer)

	return container.NewScroll(mainContent)
}
