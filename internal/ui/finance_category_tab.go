package ui

import (
	"context"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/ui/components"
	"github.com/nelsonmarro/accountable-holo/internal/ui/components/category"
)

var pageSizeOpts = []string{"10", "20", "50", "100"}

func (ui *UI) makeCategoryUI() fyne.CanvasObject {
	// Title
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Categorias",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText,
			Alignment: fyne.TextAlignCenter,
		},
	})

	// Search Bar
	searchBar := componets.NewSearchBar(ui.filterCategories)

	// Pagination and List
	ui.categoryPaginator = componets.NewPagination(
		func() (totalCount int) {
			return int(ui.categories.TotalCount)
		},
		ui.loadCategories,
		pageSizeOpts...,
	)

	ui.categoryList = widget.NewList(
		func() int {
			return len(ui.categories.Data)
		}, ui.makeCategoryListUI, ui.fillCategoryListData,
	)
	go ui.loadCategories(1, ui.categoryPaginator.GetPageSize())

	// Add Category Button
	catAddBtn := widget.NewButtonWithIcon("Agregar Categoría", theme.ContentAddIcon(), func() {
		dialogHandler := category.NewAddCategoryDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.CatService,
			func() {
				ui.loadCategories(1, ui.categoryPaginator.GetPageSize())
			},
		)

		dialogHandler.Show()
	})
	catAddBtn.Importance = widget.HighImportance

	// Containers
	titleContainer := container.NewVBox(
		container.NewCenter(title),
		container.NewBorder(nil, nil, catAddBtn, nil, searchBar),
	)

	tableHeader := container.NewBorder(
		ui.categoryPaginator,
		nil, nil, nil,
		container.NewGridWithColumns(3,
			widget.NewLabel("Nombre"),
			widget.NewLabel("Tipo"),
			widget.NewLabel("Acciones"),
		),
	)

	tableContainer := container.NewBorder(
		tableHeader, nil, nil, nil,
		ui.categoryList,
	)

	mainContent := container.NewBorder(
		container.NewPadded(titleContainer),
		nil, nil, nil,
		tableContainer,
	)

	return mainContent
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
	cat := ui.categories.Data[i]

	rowContainer := o.(*fyne.Container)

	nameLabel := rowContainer.Objects[0].(*widget.Label)
	nameLabel.SetText(cat.Name)

	typeLabel := rowContainer.Objects[1].(*widget.Label)
	typeLabel.SetText(string(cat.Type))

	actionsContainer := rowContainer.Objects[2].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	editBtn.Enable()

	deleteBtn := actionsContainer.Objects[1].(*widget.Button)
	editBtn.Enable()

	if strings.Contains(cat.Name, "Anular") {
		editBtn.Disable()
		deleteBtn.Disable()
	}

	editBtn.OnTapped = func() {
		dialogHandler := category.NewEditCategoryDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.CatService,
			func() {
				ui.loadCategories(1, ui.categoryPaginator.GetPageSize())
			},
			cat.ID,
		)

		dialogHandler.Show()
	}
	deleteBtn.OnTapped = func() {
		dialogHandler := category.NewDeleteCategoryDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.CatService,
			func() {
				ui.loadCategories(1, ui.categoryPaginator.GetPageSize())
			},
			cat.ID,
		)

		dialogHandler.Show()
	}
}

func (ui *UI) loadCategories(page int, pageSize int) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := ui.Services.CatService.GetPaginatedCategories(ctx, page, pageSize, ui.categoryFilter)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.categories = result

	fyne.Do(func() {
		ui.categoryList.Refresh()
		ui.categoryPaginator.Refresh()
	})
}

func (ui *UI) filterCategories(filter string) {
	ui.categoryFilter = filter
	ui.loadCategories(1, ui.categoryPaginator.GetPageSize())
}
