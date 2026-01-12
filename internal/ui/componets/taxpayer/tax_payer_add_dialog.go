package taxpayer

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type AddTaxPayerDialog struct {
	window   fyne.Window
	service  TaxPayerService
	onSaved  func()
}

func NewAddTaxPayerDialog(
	parent fyne.Window,
	service TaxPayerService,
	onSaved func(),
) *AddTaxPayerDialog {
	return &AddTaxPayerDialog{
		window:  parent,
		service: service,
		onSaved: onSaved,
	}
}

func (d *AddTaxPayerDialog) Show() {
	form := NewTaxPayerForm()

	dlg := dialog.NewCustomConfirm("Nuevo Cliente", "Guardar", "Cancelar", form.FormWidget, func(confirm bool) {
		if !confirm {
			return
		}

		tp := form.GetTaxPayer()
		
		// Simple validation
		if tp.Identification == "" || tp.Name == "" || tp.Email == "" {
			dialog.ShowError(fmt.Errorf("identificación, nombre y email son obligatorios"), d.window)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := d.service.Create(ctx, tp); err != nil {
			dialog.ShowError(err, d.window)
			return
		}

		if d.onSaved != nil {
			d.onSaved()
		}
		dialog.ShowInformation("Éxito", "Cliente creado correctamente", d.window)
	}, d.window)

	dlg.Resize(fyne.NewSize(400, 400))
	dlg.Show()
}
