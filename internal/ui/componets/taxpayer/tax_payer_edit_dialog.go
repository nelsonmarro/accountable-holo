package taxpayer

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type EditTaxPayerDialog struct {
	window   fyne.Window
	service  TaxPayerService
	onUpdated func()
	existing *domain.TaxPayer
}

func NewEditTaxPayerDialog(
	parent fyne.Window,
	service TaxPayerService,
	existing *domain.TaxPayer,
	onUpdated func(),
) *EditTaxPayerDialog {
	return &EditTaxPayerDialog{
		window:   parent,
		service:  service,
		existing: existing,
		onUpdated: onUpdated,
	}
}

func (d *EditTaxPayerDialog) Show() {
	form := NewTaxPayerForm()
	form.LoadData(d.existing)

	dlg := dialog.NewCustomConfirm("Editar Cliente", "Guardar", "Cancelar", form.FormWidget, func(confirm bool) {
		if !confirm {
			return
		}

		tp := form.GetTaxPayer()
		tp.ID = d.existing.ID
		// Identification shouldn't change ideally, or handle logic in service
		tp.IdentificationType = d.existing.IdentificationType 

		if tp.Name == "" || tp.Email == "" {
			dialog.ShowError(fmt.Errorf("nombre y email son obligatorios"), d.window)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := d.service.Update(ctx, tp); err != nil {
			dialog.ShowError(err, d.window)
			return
		}

		if d.onUpdated != nil {
			d.onUpdated()
		}
		dialog.ShowInformation("Éxito", "Cliente actualizado correctamente", d.window)
	}, d.window)

	dlg.Resize(fyne.NewSize(400, 400))
	dlg.Show()
}
