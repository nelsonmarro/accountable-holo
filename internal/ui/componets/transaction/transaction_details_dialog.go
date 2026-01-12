package transaction

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
)

type DetailsDialog struct {
	parent     fyne.Window
	tx         *domain.Transaction
	txService  TransactionService
	sriService SriService
}

func NewDetailsDialog(
	parent fyne.Window,
	tx *domain.Transaction,
	txService TransactionService,
	sriService SriService,
) *DetailsDialog {
	return &DetailsDialog{
		parent:     parent,
		tx:         tx,
		txService:  txService,
		sriService: sriService,
	}
}

func (d *DetailsDialog) Show() {
	progress := dialog.NewCustomWithoutButtons("Cargando detalles...", widget.NewProgressBarInfinite(), d.parent)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Cargar ítems desde la base de datos
		items, err := d.txService.GetItemsByTransactionID(ctx, d.tx.ID)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(err, d.parent)
			})
			return
		}
		d.tx.Items = items

		fyne.Do(func() {
			progress.Hide()
			content := d.buildContent()
			dlg := dialog.NewCustom("Comprobante de Transacción", "Cerrar", content, d.parent)
			dlg.Resize(fyne.NewSize(650, 600))
			dlg.Show()
		})
	}()
}

func (d *DetailsDialog) buildContent() fyne.CanvasObject {
	header := widget.NewForm(
		widget.NewFormItem("Nro. Comprobante:", widget.NewLabelWithStyle(d.tx.TransactionNumber, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewFormItem("Fecha:", widget.NewLabel(d.tx.TransactionDate.Format("02/01/2006"))),
		widget.NewFormItem("Categoría:", widget.NewLabel(d.tx.Category.Name)),
	)

	itemsContainer := container.NewVBox()
	itemsContainer.Add(container.NewGridWithColumns(4,
		widget.NewLabelWithStyle("Descripción", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Cant.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("P. Unit", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Subtotal", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
	))

	for _, item := range d.tx.Items {
		itemsContainer.Add(container.NewGridWithColumns(4,
			widget.NewLabel(item.Description),
			widget.NewLabel(fmt.Sprintf("%.2f", item.Quantity)),
			widget.NewLabel(fmt.Sprintf("$%.2f", item.UnitPrice)),
			widget.NewLabel(fmt.Sprintf("$%.2f", item.Subtotal)),
		))
	}

	footer := widget.NewForm(
		widget.NewFormItem("Subtotal 15%:", widget.NewLabel(fmt.Sprintf("$%.2f", d.tx.Subtotal15))),
		widget.NewFormItem("Subtotal 0%:", widget.NewLabel(fmt.Sprintf("$%.2f", d.tx.Subtotal0))),
		widget.NewFormItem("IVA 15%:", widget.NewLabel(fmt.Sprintf("$%.2f", d.tx.TaxAmount))),
		widget.NewFormItem("TOTAL:", widget.NewLabelWithStyle(fmt.Sprintf("$%.2f", d.tx.Amount), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
	)

	actions := container.NewHBox()
	if d.tx.ElectronicReceipt != nil && d.tx.ElectronicReceipt.SRIStatus == "AUTORIZADO" {
		statusLabel := widget.NewLabelWithStyle("✅ Factura Autorizada", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		rideBtn := widget.NewButtonWithIcon("Ver RIDE (PDF)", theme.DocumentIcon(), func() {
			d.generateAndShowRide()
		})
		actions.Add(statusLabel)
		actions.Add(rideBtn)
	} else if d.tx.ElectronicReceipt != nil {
		statusText := fmt.Sprintf("Estado SRI: %s", d.tx.ElectronicReceipt.SRIStatus)
		if d.tx.ElectronicReceipt.SRIMessage != "" {
			statusText += "\n" + d.tx.ElectronicReceipt.SRIMessage
		}
		statusLabel := widget.NewLabelWithStyle(statusText, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
		retryBtn := widget.NewButtonWithIcon("Sincronizar / Reintentar", theme.ViewRefreshIcon(), func() {
			d.retryEmission()
		})
		actions.Add(statusLabel)
		actions.Add(retryBtn)
	} else {
		emitBtn := widget.NewButtonWithIcon("Emitir Factura Electrónica", theme.ConfirmIcon(), func() {
			d.promptPasswordAndEmit()
		})
		emitBtn.Importance = widget.HighImportance
		actions.Add(emitBtn)
	}

	return container.NewVScroll(container.NewVBox(
		header,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Detalle de Ítems", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		container.NewPadded(itemsContainer),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, nil, footer),
		widget.NewSeparator(),
		container.NewCenter(actions),
	))
}

func (d *DetailsDialog) promptPasswordAndEmit() {
	passEntry := widget.NewPasswordEntry()
	items := []*widget.FormItem{
		widget.NewFormItem("Contraseña Firma:", passEntry),
	}

	formDlg := dialog.NewForm("Seguridad", "Emitir", "Cancelar", items, func(confirm bool) {
		if !confirm {
			return
		}
		d.emitFactura(passEntry.Text)
	}, d.parent)

	formDlg.Resize(fyne.NewSize(450, 200))
	formDlg.Show()
}

func (d *DetailsDialog) emitFactura(password string) {
	componets.HandleLongRunningOperation(d.parent, "Emitiendo Factura al SRI...", func(ctx context.Context) error {
		return d.sriService.EmitirFactura(ctx, d.tx.ID, password)
	})
	dialog.ShowInformation("Proceso Finalizado", "Revise el estado en la lista de transacciones.", d.parent)
}

func (d *DetailsDialog) generateAndShowRide() {
	progress := dialog.NewCustomWithoutButtons("Generando RIDE...", widget.NewProgressBarInfinite(), d.parent)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		path, err := d.sriService.GenerateRide(ctx, d.tx.ID)
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(err, d.parent)
				return
			}
			preview := NewPreviewDialog(d.parent, path)
			preview.Show()
		})
	}()
}

func (d *DetailsDialog) retryEmission() {
	progress := dialog.NewCustomWithoutButtons("Sincronizando con SRI...", widget.NewProgressBarInfinite(), d.parent)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		status, err := d.sriService.SyncReceipt(ctx, d.tx.ElectronicReceipt)
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("error de sincronización: %w", err), d.parent)
			} else {
				dialog.ShowInformation("Estado Actualizado", fmt.Sprintf("Nuevo estado: %s", status), d.parent)
			}
		})
	}()
}
