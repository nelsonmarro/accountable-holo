package ui

import (
	"context"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
			SizeName:  theme.SizeNameHeadingText,
			Alignment: fyne.TextAlignCenter,
		},
	})

	catAddBtn := widget.NewButtonWithIcon("Agregar Categoría", theme.ContentAddIcon(), func() {})
	catAddBtn.Importance = widget.HighImportance

	ui.categories = []domain.Category{
		{BaseEntity: domain.BaseEntity{ID: 1}, Name: "Sales Revenue", Type: "income"},
		{BaseEntity: domain.BaseEntity{ID: 2}, Name: "Office Supplies", Type: "outcome"},
		{BaseEntity: domain.BaseEntity{ID: 3}, Name: "Consulting Services", Type: "income"},
		{BaseEntity: domain.BaseEntity{ID: 4}, Name: "Software Subscriptions", Type: "outcome"},
	}

	header := container.NewGridWithColumns(3,
		widget.NewLabel("Name"),
		widget.NewLabel("Type"),
		widget.NewLabel("Actions"),
	)

	ui.categoryList = widget.NewList(
		func() int {
			return len(ui.categories)
		}, ui.makeCategoryListUI, ui.fillCategoryListData,
	)
	go ui.loadCategories()

	// containers
	headerContainer := container.NewVBox(
		container.NewCenter(title),
		container.NewHBox(layout.NewSpacer(), catAddBtn),
	)
	tableContainer := container.NewBorder(header, nil, nil, nil, ui.categoryList)
	mainContent := container.NewBorder(container.NewPadded(headerContainer), nil, nil, nil, tableContainer)

	return container.NewScroll(mainContent)
}

func (ui *UI) makeCategoryListUI() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editBtn.Importance = widget.HighImportance

	delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	delBtn.Importance = widget.DangerImportance

	return container.NewGridWithColumns(3,
		widget.NewLabel("template name"),
		widget.NewLabel("template type"),
		container.NewHBox( // A container for our action buttons
			editBtn,
			delBtn,
		),
	)
}

func (ui *UI) fillCategoryListData(i widget.ListItemID, o fyne.CanvasObject) {
	category := ui.categories[i]

	rowContainer := o.(*fyne.Container)

	nameLabel := rowContainer.Objects[0].(*widget.Label)
	nameLabel.SetText(category.Name)

	typeLabel := rowContainer.Objects[1].(*widget.Label)
	typeLabel.SetText(string(category.Type))

	actionsContainer := rowContainer.Objects[2].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	deleteBtn := actionsContainer.Objects[1].(*widget.Button)

	editBtn.OnTapped = func() {
		log.Printf("Edit button tapped for category ID: %d, Name: %s", category.ID, category.Name)
	}
	deleteBtn.OnTapped = func() {
		log.Printf("Delete button tapped for category ID: %d, Name: %s", category.ID, category.Name)
	}
}

func (ui *UI) loadCategories() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := ui.catService.GetPaginatedCategories(ctx, 1, 2)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.categories = result.Data

	fyne.Do(func() {
		ui.categoryList.Refresh()
	})
}
