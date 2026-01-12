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
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TaxPayerService interface {
	Create(ctx context.Context, tp *domain.TaxPayer) error
	Search(ctx context.Context, query string) ([]domain.TaxPayer, error)
	GetByIdentification(ctx context.Context, identification string) (*domain.TaxPayer, error)
}

type SearchDialog struct {
	window     fyne.Window
	logger     *log.Logger
	service    TaxPayerService
	onSelected func(*domain.TaxPayer)
}

func NewSearchDialog(
	parent fyne.Window,
	logger *log.Logger,
	service TaxPayerService,
	onSelected func(*domain.TaxPayer),
) *SearchDialog {
	return &SearchDialog{
		window:     parent,
		logger:     logger,
		service:    service,
		onSelected: onSelected,
	}
}

func (d *SearchDialog) Show() {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Buscar por RUC o Nombre...")

	resultsList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {},
	)

	var searchResults []domain.TaxPayer

	// Search Logic
	doSearch := func(query string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		results, err := d.service.Search(ctx, query)
		if err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		
		// Client-side filtering if API returns all
		var filtered []domain.TaxPayer
		for _, tp := range results {
			// Simple contains check (case sensitive for now, improve later)
			if query == "" || 
			   (len(tp.Identification) >= len(query) && tp.Identification[:len(query)] == query) ||
			   (len(tp.Name) >= len(query) && tp.Name[:len(query)] == query) {
				filtered = append(filtered, tp)
			}
		}
		
		searchResults = filtered
		resultsList.Length = func() int { return len(searchResults) }
		resultsList.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			tp := searchResults[i]
			o.(*widget.Label).SetText(fmt.Sprintf("%s - %s", tp.Identification, tp.Name))
		}
		resultsList.Refresh()
	}

	searchEntry.OnChanged = doSearch

	// Selection Logic
	var dlg dialog.Dialog
	resultsList.OnSelected = func(id widget.ListItemID) {
		d.onSelected(&searchResults[id])
		dlg.Hide()
	}

	// Create New Logic
	createBtn := widget.NewButtonWithIcon("Nuevo Cliente", theme.ContentAddIcon(), func() {
		d.showCreateDialog()
		dlg.Hide()
	})

	content := container.NewBorder(
		container.NewVBox(searchEntry, createBtn), 
		nil, nil, nil, 
		resultsList,
	)

	dlg = dialog.NewCustom("Buscar Cliente", "Cerrar", content, d.window)
	dlg.Resize(fyne.NewSize(400, 500))
	
	// Initial Load
	doSearch("")
	
	dlg.Show()
}

func (d *SearchDialog) showCreateDialog() {
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

	dialog.ShowForm("Nuevo Cliente", "Guardar", "Cancelar", items, func(confirm bool) {
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

		// Auto-select created client
		// We need to fetch it again to get the ID if Create doesn't set it (Repo implementation dependent)
		// Assuming Create sets ID or we fetch by ID
		created, _ := d.service.GetByIdentification(ctx, tp.Identification)
		if created != nil {
			d.onSelected(created)
		}
	}, d.window)
}
