package transaction

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
)

type SriQueueDialog struct {
	parent     fyne.Window
	sriService SriService
	txService  TransactionService // Needed for details
	dialog     dialog.Dialog
	data       []domain.ElectronicReceipt
	list       *widget.List
}

func NewSriQueueDialog(parent fyne.Window, sriService SriService, txService TransactionService) *SriQueueDialog {
	return &SriQueueDialog{
		parent:     parent,
		sriService: sriService,
		txService:  txService,
	}
}

func (d *SriQueueDialog) Show() {
	// Contenido
	d.list = widget.NewList(
		func() int { return len(d.data) },
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(5,
				widget.NewLabel("01/01 15:00"),
				widget.NewLabel("001-001-000000001"),
				widget.NewLabel("Cliente Nombre"),
				widget.NewLabel("$100.00"),
				container.NewHBox(widget.NewIcon(theme.HistoryIcon()), widget.NewLabel("ESTADO")),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(d.data) {
				return
			}
			r := d.data[i]
			box := o.(*fyne.Container)
			
			// Date
			box.Objects[0].(*widget.Label).SetText(r.CreatedAt.Format("02/01 15:04"))
			
			// Nro Comprobante
			box.Objects[1].(*widget.Label).SetText(r.TransactionNumber)

			// Cliente
			box.Objects[2].(*widget.Label).SetText(r.ClientName)
			box.Objects[2].(*widget.Label).Truncation = fyne.TextTruncateEllipsis

			// Total
			box.Objects[3].(*widget.Label).SetText(fmt.Sprintf("$%.2f", r.TotalAmount))

			// Status + Icon
			statusBox := box.Objects[4].(*fyne.Container)
			icon := statusBox.Objects[0].(*widget.Icon)
			lbl := statusBox.Objects[1].(*widget.Label)
			
			lbl.SetText(r.SRIStatus)
			switch r.SRIStatus {
			case "RECIBIDA":
				icon.SetResource(theme.ConfirmIcon())
			case "EN PROCESO":
				icon.SetResource(theme.HistoryIcon()) // Reloj
			case "ERROR_RED", "ERROR_ENVIO":
				icon.SetResource(theme.WarningIcon())
			default: // PENDIENTE
				icon.SetResource(theme.QuestionIcon())
			}
		},
	)

	d.list.OnSelected = func(id widget.ListItemID) {
		r := d.data[id]
		// Abrir detalles
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			tx, err := d.txService.GetTransactionByID(ctx, r.TransactionID)
			if err != nil {
				fyne.Do(func() { dialog.ShowError(err, d.parent) })
				return
			}
			// Inject receipt details manually to ensure latest status
			tx.ElectronicReceipt = &r

			fyne.Do(func() {
				detailsDlg := NewDetailsDialog(d.parent, tx, d.txService, d.sriService, func() {
					d.loadData() // Recargar lista al cerrar detalles
				})
				detailsDlg.Show()
				d.list.Unselect(id)
			})
		}()
	}

	header := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Fecha", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Nro. Doc", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Cliente", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Total", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Estado SRI", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	refreshBtn := widget.NewButtonWithIcon("Recargar", theme.ViewRefreshIcon(), func() {
		d.loadData()
	})

	processBtn := widget.NewButtonWithIcon("Forzar Sincronización", theme.MediaPlayIcon(), func() {
		d.forceSync()
	})
	processBtn.Importance = widget.HighImportance

	infoLabel := widget.NewLabel("El sistema verifica automáticamente cada 5 minutos.")
	infoLabel.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(refreshBtn, processBtn, layout.NewSpacer(), infoLabel),
			header,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		d.list,
	)

	d.dialog = dialog.NewCustom("Cola de Procesamiento SRI", "Cerrar", content, d.parent)
	d.dialog.Resize(fyne.NewSize(800, 500))
	
	// Cargar datos iniciales
	d.loadData()
	
	d.dialog.Show()
}

func (d *SriQueueDialog) loadData() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		receipts, err := d.sriService.GetPendingQueue(ctx)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, d.parent) })
			return
		}
		
		d.data = receipts
		fyne.Do(func() { d.list.Refresh() })
	}()
}

func (d *SriQueueDialog) forceSync() {
	var count int
	componets.HandleLongRunningOperation(d.parent, "Procesando Cola SRI...", func(ctx context.Context) error {
		var err error
		count, err = d.sriService.ProcessBackgroundSync(ctx)
		return err
	}, func() {
		dialog.ShowInformation("Sincronización Completa", fmt.Sprintf("Se procesaron %d comprobantes.", count), d.parent)
		d.loadData() // Recargar lista
	})
}
