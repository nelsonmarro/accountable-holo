package transaction

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
)

type DetailsDialog struct {
	parent     fyne.Window
	tx         *domain.Transaction
	txService  TransactionService
	sriService SriService
	onChanged  func()
	dialog     dialog.Dialog // Added reference
}

func NewDetailsDialog(
	parent fyne.Window,
	tx *domain.Transaction,
	txService TransactionService,
	sriService SriService,
	onChanged func(), // Added
) *DetailsDialog {
	return &DetailsDialog{
		parent:     parent,
		tx:         tx,
		txService:  txService,
		sriService: sriService,
		onChanged:  onChanged, // Added
	}
}

func (d *DetailsDialog) Show() {
	progress := dialog.NewCustomWithoutButtons("Cargando detalles...", widget.NewProgressBarInfinite(), d.parent)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 1. Recargar Transacción Completa (CRÍTICO para tener ReceiptType y estado actualizado)
		fullTx, err := d.txService.GetTransactionByID(ctx, d.tx.ID)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("error cargando datos de transacción: %w", err), d.parent)
			})
			return
		}
		d.tx = fullTx // Actualizar referencia

		// 2. Cargar ítems
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
			d.dialog = dialog.NewCustom("Comprobante de Transacción", "Cerrar", content, d.parent)
			d.dialog.Resize(fyne.NewSize(750, 650))
			d.dialog.Show()
		})
	}()
}

func (d *DetailsDialog) buildContent() fyne.CanvasObject {
	var catTypeColor color.Color
	if d.tx.Category.Type == domain.Income {
		catTypeColor = color.NRGBA{R: 0, G: 150, B: 0, A: 255}
	} else {
		catTypeColor = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
	}

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetText(d.tx.Description)
	header := widget.NewForm(
		widget.NewFormItem("Nro. Comprobante:", widget.NewLabelWithStyle(d.tx.TransactionNumber, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewFormItem("Fecha:", widget.NewLabel(d.tx.TransactionDate.Format(componets.AppDateFormat))),
		widget.NewFormItem("Descripción", descriptionEntry),
		widget.NewFormItem("Categoría:", container.NewHBox(
			widget.NewLabel(fmt.Sprintf("%s   -", d.tx.Category.Name)),
			canvas.NewText(string(d.tx.Category.Type), catTypeColor),
		)),
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

	// Determine SRI Action Button
	// Show if:
	// 1. It already has an electronic receipt (Invoice or Credit Note).
	// 2. OR it is a pure Income (Sale) that is NOT a reversal/void of another transaction.
	if (d.tx.Category.Type == domain.Income && d.tx.VoidsTransactionID == nil) || d.tx.ElectronicReceipt != nil {
		isAuthorized := d.tx.ElectronicReceipt != nil && d.tx.ElectronicReceipt.SRIStatus == "AUTORIZADO"
		hasReceipt := d.tx.ElectronicReceipt != nil

		if isAuthorized {
			statusLabel := widget.NewLabelWithStyle("✅ Comprobante Autorizado", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			rideBtn := widget.NewButtonWithIcon("Ver RIDE (PDF)", theme.DocumentIcon(), func() {
				d.generateAndShowRide()
			})
			actions.Add(statusLabel)
			actions.Add(rideBtn)

			// Email Status logic
			if d.tx.ElectronicReceipt.EmailSent {
				emailLabel := widget.NewLabelWithStyle("✉️ Email Enviado", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
				actions.Add(emailLabel)
			} else {
				resendBtn := widget.NewButtonWithIcon("Reenviar Email", theme.MailSendIcon(), func() {
					d.resendEmail()
				})
				resendBtn.Importance = widget.WarningImportance
				actions.Add(resendBtn)
			}

		} else if hasReceipt {
			status := d.tx.ElectronicReceipt.SRIStatus
			statusText := fmt.Sprintf("Estado SRI: %s", status)

			// Detectar si está "trabada" en proceso por mucho tiempo (> 2 horas)
			isStuck := status == "EN PROCESO" && time.Since(d.tx.ElectronicReceipt.CreatedAt) > 2*time.Hour

			// Si es error temporal o está en proceso reciente, permitir sincronizar
			if (status == "RECIBIDA" || status == "EN PROCESO" || status == "PENDIENTE" || status == "ERROR_RED") && !isStuck {
				statusLabel := widget.NewLabelWithStyle(statusText, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
				retryBtn := widget.NewButtonWithIcon("Sincronizar / Reintentar", theme.ViewRefreshIcon(), func() {
					d.retryEmission()
				})
				actions.Add(statusLabel)
				actions.Add(retryBtn)
			} else {
				// Terminal errors (NO AUTORIZADO, RECHAZADA, DEVUELTA) OR Stuck Process: Allow RE-EMIT (New Key)
				label := statusText + " (Fallido)"
				if isStuck {
					label = statusText + " (Sin respuesta > 2h)"
				}
				statusLabel := widget.NewLabelWithStyle(label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
				reEmitBtn := widget.NewButtonWithIcon("Corregir y Re-Emitir", theme.ConfirmIcon(), func() {
					d.promptPasswordAndEmit()
				})
				reEmitBtn.Importance = widget.WarningImportance
				actions.Add(statusLabel)
				actions.Add(reEmitBtn)
			}
		} else {
			// New emission (Only for Income, don't allow creating receipt for arbitrary outcomes unless via void process)
			if d.tx.Category.Type == domain.Income {
				emitBtn := widget.NewButtonWithIcon("Emitir Factura Electrónica", theme.ConfirmIcon(), func() {
					d.promptPasswordAndEmit()
				})
				emitBtn.Importance = widget.HighImportance
				actions.Add(emitBtn)
			}
		}
	}

	// Layout principal: Acciones arriba, detalles abajo con scroll
	var topActions fyne.CanvasObject
	if len(actions.Objects) > 0 {
		topActions = container.NewVBox(
			container.NewCenter(actions),
			widget.NewSeparator(),
		)
	}

	detailsContent := container.NewVBox(
		header,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Detalle de Ítems", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		container.NewPadded(itemsContainer),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, nil, footer),
	)

	return container.NewBorder(
		topActions,
		nil, nil, nil,
		container.NewVScroll(detailsContent),
	)
}

func (d *DetailsDialog) promptPasswordAndEmit() {
	passEntry := widget.NewPasswordEntry()
	motivoEntry := widget.NewEntry()
	motivoEntry.SetPlaceHolder("Razón de la corrección")

	items := []*widget.FormItem{
		widget.NewFormItem("Contraseña Firma:", passEntry),
	}

	// Si es Nota de Crédito (Tipo 04), pedir motivo
	// Es NC si el recibo dice "04" O si la transacción anula a otra (VoidsTransactionID != nil)
	isNC := (d.tx.ElectronicReceipt != nil && d.tx.ElectronicReceipt.ReceiptType == "04") || d.tx.VoidsTransactionID != nil
	if isNC {
		items = append(items, widget.NewFormItem("Motivo (NC):", motivoEntry))
	}

	formDlg := dialog.NewForm("Seguridad", "Emitir", "Cancelar", items, func(confirm bool) {
		if !confirm {
			return
		}

		motivo := "Corrección"
		if isNC && motivoEntry.Text != "" {
			motivo = motivoEntry.Text
		}
		d.emitDocument(passEntry.Text, motivo)
	}, d.parent)

	formDlg.Resize(fyne.NewSize(450, 200))
	formDlg.Show()
}

func (d *DetailsDialog) emitDocument(password string, motivo string) {
	msg := "Emitiendo Factura al SRI..."
	// Es NC si el recibo dice "04" O si la transacción anula a otra
	isNC := (d.tx.ElectronicReceipt != nil && d.tx.ElectronicReceipt.ReceiptType == "04") || d.tx.VoidsTransactionID != nil

	if isNC {
		msg = "Emitiendo Nota de Crédito al SRI..."
	}

	componets.HandleLongRunningOperation(d.parent, msg, func(ctx context.Context) error {
		var err error

		// Decidir qué emitir
		if isNC {
			// Es Nota de Crédito
			if d.tx.VoidsTransactionID == nil {
				return errors.New("error de datos: esta transacción de anulación no está vinculada a una factura original")
			}
			_, err = d.sriService.EmitirNotaCredito(ctx, d.tx.ID, *d.tx.VoidsTransactionID, motivo, password)
		} else {
			// Es Factura (Por defecto o Tipo 01)
			err = d.sriService.EmitirFactura(ctx, d.tx.ID, password)
		}
		if err != nil {
			return err
		}

		// Éxito técnico (se envió). Ahora verificamos el estado de negocio.
		// Recargamos la transacción para ver cómo quedó.
		updatedTx, errFetch := d.txService.GetTransactionByID(ctx, d.tx.ID)
		if errFetch != nil {
			return fmt.Errorf("documento enviado, pero error al recargar estado: %w", errFetch)
		}

		fyne.Do(func() {
			if d.dialog != nil {
				d.dialog.Hide() // Cerrar detalles
			}

			d.showSRIFeedback(updatedTx) // Mostrar feedback rico

			if d.onChanged != nil {
				d.onChanged()
			}
		})
		return nil
	}, nil)
}

func (d *DetailsDialog) showSRIFeedback(tx *domain.Transaction) {
	if tx.ElectronicReceipt == nil {
		return
	}

	status := tx.ElectronicReceipt.SRIStatus
	var title, message string
	var icon fyne.Resource

	switch status {
	case "AUTORIZADO":
		title = "✅ Factura Autorizada"
		message = fmt.Sprintf("La factura ha sido autorizada correctamente.\nClave: %s", tx.ElectronicReceipt.AccessKey)
		icon = theme.ConfirmIcon()
	case "EN PROCESO", "RECIBIDA", "PENDIENTE":
		title = "⏳ En Proceso"
		message = "La factura fue recibida por el SRI pero la autorización está pendiente.\nEl sistema verificará el estado automáticamente."
		icon = theme.InfoIcon()
	case "NO AUTORIZADO", "RECHAZADA", "DEVUELTA":
		title = "❌ Factura Rechazada"
		message = fmt.Sprintf("El SRI rechazó el comprobante.\nMotivo: %s", tx.ElectronicReceipt.SRIMessage)
		icon = theme.ErrorIcon()
	default:
		title = "Estado Desconocido"
		message = fmt.Sprintf("Estado actual: %s", status)
		icon = theme.QuestionIcon()
	}

	content := container.NewHBox(
		widget.NewIcon(icon),
		widget.NewLabel(message),
	)

	dialog.ShowCustom(title, "Aceptar", content, d.parent)
}

func (d *DetailsDialog) generateAndShowRide() {
	progress := dialog.NewCustomWithoutButtons("Generando RIDE...", widget.NewProgressBarInfinite(), d.parent)
	progress.Show()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 1. Generar en TEMP
		tempPath, err := d.sriService.GenerateRide(ctx, d.tx.ID)

		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(err, d.parent)
				return
			}

			// 2. Diálogo Guardar Como
			saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				defer os.Remove(tempPath)

				if err != nil {
					dialog.ShowError(err, d.parent)
					return
				}
				if writer == nil {
					return
				}
				defer writer.Close()

				// 3. Copiar contenido: Temp -> Destino Usuario
				srcFile, err := os.Open(tempPath)
				if err != nil {
					dialog.ShowError(fmt.Errorf("error leyendo archivo temporal: %w", err), d.parent)
					return
				}
				defer srcFile.Close()

				if _, err := io.Copy(writer, srcFile); err != nil {
					dialog.ShowError(fmt.Errorf("error guardando archivo: %w", err), d.parent)
					return
				}

				dialog.ShowInformation("Éxito", "RIDE guardado correctamente", d.parent)
			}, d.parent)

			saveDialog.SetFileName(fmt.Sprintf("RIDE-%s.pdf", d.tx.ElectronicReceipt.AccessKey))
			saveDialog.Show()
		})
	}()
}

func (d *DetailsDialog) resendEmail() {
	componets.HandleLongRunningOperation(d.parent, "Reenviando Email...", func(ctx context.Context) error {
		return d.sriService.ResendEmail(ctx, d.tx.ID)
	}, func() {
		dialog.ShowInformation("Éxito", "El correo ha sido enviado exitosamente.", d.parent)

		// Forzar cierre del diálogo para evitar re-envíos y refrescar estado al reabrir
		if d.dialog != nil {
			d.dialog.Hide()
		}

		// Recargar estado para actualizar lista padre
		if d.onChanged != nil {
			d.onChanged()
		}
	})
}

func (d *DetailsDialog) retryEmission() {
	componets.HandleLongRunningOperation(d.parent, "Sincronizando con SRI...", func(ctx context.Context) error {
		_, err := d.sriService.SyncReceipt(ctx, d.tx.ElectronicReceipt)
		if err != nil {
			return fmt.Errorf("error de sincronización: %w", err)
		}

		// Recargar estado
		updatedTx, errFetch := d.txService.GetTransactionByID(ctx, d.tx.ID)
		if errFetch != nil {
			return fmt.Errorf("sincronizado, pero error al recargar: %w", errFetch)
		}

		fyne.Do(func() {
			if d.dialog != nil {
				d.dialog.Hide()
			}

			d.showSRIFeedback(updatedTx)

			if d.onChanged != nil {
				d.onChanged()
			}
		})
		return nil
	}, nil)
}
