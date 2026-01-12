package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/taxpayer"
)

func (ui *UI) makeTaxPayerTab() fyne.CanvasObject {
	// Title
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Gestión de Clientes",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText,
			Alignment: fyne.TextAlignCenter,
		},
	})

	searchEntry := componets.NewSearchBar(func(s string) {
		ui.taxPayerSearchText = s
		go ui.loadTaxPayers(1)
	})

	ui.taxPayerPaginator = componets.NewPagination(
		func() int {
			if ui.paginatedTaxPayers == nil {
				return 0
			}
			return int(ui.paginatedTaxPayers.TotalCount)
		},
		func(page, pageSize int) {
			go ui.loadTaxPayers(page)
		},
		pageSizeOpts...,
	)

	ui.taxPayerList = widget.NewList(
		func() int {
			if ui.paginatedTaxPayers == nil {
				return 0
			}
			return len(ui.paginatedTaxPayers.Data)
		},
		ui.makeTaxPayerListUI,
		ui.fillTaxPayerListData,
	)

	addBtn := widget.NewButtonWithIcon("Nuevo Cliente", theme.ContentAddIcon(), func() {
		dlg := taxpayer.NewAddTaxPayerDialog(
			ui.mainWindow,
			ui.Services.TaxService,
			func() {
				go ui.loadTaxPayers(1)
			},
		)
		dlg.Show()
	})
	addBtn.Importance = widget.HighImportance

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		go ui.loadTaxPayers(1)
	})

	topBar := container.NewBorder(nil, nil,
		container.NewHBox(addBtn, refreshBtn),
		nil,
		searchEntry,
	)

	// Table Header
	header := container.NewGridWithColumns(4,
		widget.NewLabelWithStyle("Nombre", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Email", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Acción", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	tableContainer := container.NewBorder(header, nil, nil, nil, ui.taxPayerList)

	content := container.NewBorder(
		container.NewVBox(container.NewCenter(title), topBar),
		ui.taxPayerPaginator, nil, nil,
		tableContainer,
	)

	go ui.loadTaxPayers(1)

	return content
}

func (ui *UI) makeTaxPayerListUI() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	// No delete button as per requirement

	return container.NewGridWithColumns(4,
		widget.NewLabel("Template Name"),
		widget.NewLabel("Template ID"),
		widget.NewLabel("Template Email"),
		container.NewHBox(editBtn),
	)
}

func (ui *UI) fillTaxPayerListData(i widget.ListItemID, o fyne.CanvasObject) {
	if ui.paginatedTaxPayers == nil || i >= len(ui.paginatedTaxPayers.Data) {
		return
	}
	tp := ui.paginatedTaxPayers.Data[i]

	box := o.(*fyne.Container)
	box.Objects[0].(*widget.Label).SetText(tp.Name)
	box.Objects[1].(*widget.Label).SetText(tp.Identification)
	box.Objects[2].(*widget.Label).SetText(tp.Email)

	actionsBox := box.Objects[3].(*fyne.Container)
	editBtn := actionsBox.Objects[0].(*widget.Button)

	editBtn.OnTapped = func() {
		dlg := taxpayer.NewEditTaxPayerDialog(
			ui.mainWindow,
			ui.Services.TaxService,
			&tp,
			func() {
				go ui.loadTaxPayers(ui.taxPayerPaginator.GetCurrentPage())
			},
		)
		dlg.Show()
	}
}

func (ui *UI) loadTaxPayers(page int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pageSize := ui.taxPayerPaginator.GetPageSize()
	results, err := ui.Services.TaxService.GetPaginated(ctx, page, pageSize, ui.taxPayerSearchText)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error cargando clientes: %w", err), ui.mainWindow)
		})
		ui.errorLogger.Printf("Error loading taxpayers: %v", err)
		return
	}

	ui.paginatedTaxPayers = results
	fyne.Do(func() {
		ui.taxPayerList.Refresh()
		ui.taxPayerPaginator.Refresh()
	})
}
