// Package taxpayer implements a dialog for searching and selecting taxpayers.
package taxpayer

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
)

type SearchDialog struct {
	window     fyne.Window
	logger     *log.Logger
	service    TaxPayerService
	onSelected func(*domain.TaxPayer)

	// UI Components
	searchEntry *widget.Entry
	list        *widget.List
	pagination  *componets.Pagination
	dl          dialog.Dialog

	// State
	taxpayers  []domain.TaxPayer
	totalCount int64
	searchTerm string
}

var pageSizeOpts = []string{"10", "20", "50"}

func NewSearchDialog(
	parent fyne.Window,
	logger *log.Logger,
	service TaxPayerService,
	onSelected func(*domain.TaxPayer),
) *SearchDialog {
	d := &SearchDialog{
		window:     parent,
		logger:     logger,
		service:    service,
		onSelected: onSelected,
	}
	d.searchEntry = widget.NewEntry()
	d.searchEntry.SetPlaceHolder("Buscar por RUC o Nombre...")
	return d
}

func (d *SearchDialog) Show() {
	content := d.createContent()

	d.dl = dialog.NewCustom("Buscar Cliente", "Cerrar", content, d.window)
	d.dl.Resize(fyne.NewSize(750, 600))

	go d.loadTaxPayers(1, d.pagination.GetPageSize())

	d.dl.Show()
}

func (d *SearchDialog) createContent() fyne.CanvasObject {
	// --- Search Bar ---
	d.searchEntry.OnChanged = func(s string) {
		time.AfterFunc(300*time.Millisecond, func() {
			if s == d.searchEntry.Text {
				d.filterTaxPayers(s)
			}
		})
	}

	// --- Pagination ---
	d.pagination = componets.NewPagination(
		func() int { return int(d.totalCount) },
		d.loadTaxPayers,
		pageSizeOpts...,
	)

	// --- List ---
	d.list = widget.NewList(
		func() int { return len(d.taxpayers) },
		d.makeListItem,
		d.updateListItem,
	)

	// --- Headers ---
	header := container.NewGridWithColumns(3,
		widget.NewLabel("Identificación"),
		widget.NewLabel("Nombre / Razón Social"),
		widget.NewLabel("Acción"),
	)

	// --- Create Button ---
	createBtn := widget.NewButtonWithIcon("Nuevo Cliente", theme.ContentAddIcon(), func() {
		d.showCreateDialog()
	})

	// --- Layout ---
	topContainer := container.NewBorder(nil, nil, nil, createBtn, d.searchEntry)
	tableContainer := container.NewBorder(header, nil, nil, nil, d.list)

	return container.NewBorder(
		container.NewVBox(topContainer, d.pagination),
		nil, nil, nil,
		tableContainer,
	)
}

func (d *SearchDialog) makeListItem() fyne.CanvasObject {
	lblID := widget.NewLabel("0000000000001")
	lblName := widget.NewLabel("Template Name")
	selectBtn := widget.NewButtonWithIcon("Seleccionar", theme.ConfirmIcon(), nil)
	selectBtn.Importance = widget.HighImportance

	return container.NewGridWithColumns(3,
		lblID,
		lblName,
		container.NewCenter(selectBtn),
	)
}

func (d *SearchDialog) updateListItem(i widget.ListItemID, o fyne.CanvasObject) {
	if i >= len(d.taxpayers) {
		return
	}
	tp := d.taxpayers[i]
	row := o.(*fyne.Container)

	row.Objects[0].(*widget.Label).SetText(tp.Identification)
	row.Objects[1].(*widget.Label).SetText(tp.Name)

	btnContainer := row.Objects[2].(*fyne.Container)
	btn := btnContainer.Objects[0].(*widget.Button)
	btn.OnTapped = func() {
		if d.onSelected != nil {
			d.onSelected(&tp)
		}
		d.dl.Hide()
	}
}

func (d *SearchDialog) loadTaxPayers(page int, pageSize int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := d.service.GetPaginated(ctx, page, pageSize, d.searchTerm)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error cargando clientes: %w", err), d.window)
		})
		return
	}

	d.taxpayers = result.Data
	d.totalCount = result.TotalCount

	fyne.Do(func() {
		d.pagination.Refresh()
		d.list.Refresh()
	})
}

func (d *SearchDialog) filterTaxPayers(filter string) {
	d.searchTerm = filter
	d.loadTaxPayers(1, d.pagination.GetPageSize())
}

func (d *SearchDialog) showCreateDialog() {
	// ... (Misma lógica de creación, pero refresca la lista al final)
	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("17900...")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Razón Social")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("cliente@email.com")

	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("Dirección")

	items := []*widget.FormItem{
		widget.NewFormItem("Identificación (RUC/Cédula)", idEntry),
		widget.NewFormItem("Nombre / Razón Social", nameEntry),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Dirección", addrEntry),
	}

	dForm := dialog.NewForm("Nuevo Cliente", "Guardar", "Cancelar", items, func(confirm bool) {
		if !confirm {
			return
		}

		tp := &domain.TaxPayer{
			Identification:     idEntry.Text,
			IdentificationType: "04", // Default RUC
			Name:               nameEntry.Text,
			Email:              emailEntry.Text,
			Address:            addrEntry.Text,
		}

		if len(tp.Identification) == 10 {
			tp.IdentificationType = "05" // Cédula
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := d.service.Create(ctx, tp); err != nil {
			dialog.ShowError(err, d.window)
			return
		}

		// Recargar la lista y seleccionar
		d.filterTaxPayers(tp.Identification) // Busca el nuevo

		// Auto-seleccionar si se encuentra
		created, _ := d.service.GetByIdentification(ctx, tp.Identification)
		if created != nil {
			d.onSelected(created)
			d.dl.Hide()
		}
	}, d.window)

	dForm.Resize(fyne.NewSize(520, 350))
	dForm.Show()
}
