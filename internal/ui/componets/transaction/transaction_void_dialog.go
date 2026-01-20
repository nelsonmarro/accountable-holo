package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
)

// VoidTransactionDialog holds the dependencies for the delete confirmation dialog.
type VoidTransactionDialog struct {
	mainWin        fyne.Window
	logger         *log.Logger
	service        TransactionService
	sriService     SriService
	taxService     TaxPayerService // Injected to validate customer
	callbackAction func()
	txID           int
	currentUser    domain.User
}

// NewVoidTransactionDialog creates a new dialog handler for the delete action.
func NewVoidTransactionDialog(
	win fyne.Window,
	l *log.Logger,
	service TransactionService,
	sriService SriService,
	taxService TaxPayerService, // Injected
	callback func(),
	txID int,
	currentUser domain.User,
) *VoidTransactionDialog {
	return &VoidTransactionDialog{
		mainWin:        win,
		logger:         l,
		service:        service,
		sriService:     sriService,
		taxService:     taxService, // Injected
		callbackAction: callback,
		txID:           txID,
		currentUser:    currentUser,
	}
}

// Show displays the confirmation dialog to the user.
func (d *VoidTransactionDialog) Show() {
	// 1. Obtener la transaccion para decidir el flujo
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := d.service.GetTransactionByID(ctx, d.txID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error cargando transacción: %w", err), d.mainWin)
		return
	}

	// 2. Decidir flujo segun estado SRI
	isAuthorized := tx.ElectronicReceipt != nil && tx.ElectronicReceipt.SRIStatus == "AUTORIZADO"

	if isAuthorized && tx.Category.Type == domain.Income {
		d.showSRIVoidDialog(ctx, tx) // Pass context
	} else {
		d.showLocalVoidDialog()
	}
}

func (d *VoidTransactionDialog) showLocalVoidDialog() {
	confirmCallback := func(confirmed bool) {
		if !confirmed {
			return
		}
		d.executeVoid(nil)
	}

	dialog.ShowConfirm("Confirmar Anulación",
		"Está seguro que desea anular esta Transacción?. Anular la transacción creará una contrapartida interna",
		confirmCallback,
		d.mainWin,
	)
}

// Flujo SRI (Nota de Credito)
func (d *VoidTransactionDialog) showSRIVoidDialog(ctx context.Context, tx *domain.Transaction) {
	// Validacion Consumidor Final usando TaxService
	if tx.TaxPayerID != nil {
		client, err := d.taxService.GetByID(ctx, *tx.TaxPayerID)
		if err == nil && client != nil {
			if client.Identification == "9999999999999" {
				dialog.ShowError(errors.New("NORMARTIVA SRI 2026: No se pueden anular facturas de CONSUMIDOR FINAL una vez autorizadas."), d.mainWin)
				return
			}
		}
	} else {
		// Si no hay cliente vinculado en una factura autorizada, asumimos CF o error de datos
		// Pero para seguridad, bloqueamos o advertimos.
		// En este caso, asumiremos que si no hay ID, es CF implicito (como se manejaba antes).
		dialog.ShowError(errors.New("error: no se puede validar el cliente. Posible Consumidor Final."), d.mainWin)
		return
	}

	motivoEntry := widget.NewEntry()
	motivoEntry.SetPlaceHolder("Razón de la anulación (Obligatorio)")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Contraseña de Firma Electrónica (Obligatorio)")

	items := []*widget.FormItem{
		widget.NewFormItem("ATENCIÓN", widget.NewLabel("Esta factura está AUTORIZADA por el SRI.\nSe emitirá una NOTA DE CRÉDITO para anularla legalmente.")),
		widget.NewFormItem("Motivo", motivoEntry),
		widget.NewFormItem("Contraseña", passEntry),
	}

	submitFunc := func(confirm bool) {
		if !confirm {
			return
		}
		if strings.TrimSpace(motivoEntry.Text) == "" {
			dialog.ShowError(errors.New("el motivo es obligatorio para el SRI"), d.mainWin)
			return
		}
		if passEntry.Text == "" {
			dialog.ShowError(errors.New("la contraseña es obligatoria"), d.mainWin)
			return
		}

		params := &sriVoidParams{
			Motivo:       motivoEntry.Text,
			Password:     passEntry.Text,
			OriginalTxID: tx.ID,
		}
		d.executeVoid(params)
	}

	dlg := dialog.NewForm("Anular Factura SRI",
		"Emitir Nota de Crédito",
		"Cancelar",
		items,
		submitFunc,
		d.mainWin,
	)
	dlg.Resize(fyne.NewSize(500, 300))
	dlg.Show()
}

type sriVoidParams struct {
	Motivo       string
	Password     string
	OriginalTxID int
}

// executeVoid runs the actual voiding logic
func (d *VoidTransactionDialog) executeVoid(sriParams *sriVoidParams) {
	msg := "Anulando internamente..."
	if sriParams != nil {
		msg = "Emitiendo Nota de Crédito SRI..."
	}

	var accessKey string

	// Definir la operación principal
	operation := func(ctx context.Context) error {
		// Anulacion Local (Siempre se hace)
		voidTxID, err := d.service.VoidTransaction(ctx, d.txID, d.currentUser)
		if err != nil {
			return fmt.Errorf("error en la anulación local: %w", err)
		}

		// Si es SRI, emitir Nota de Credito vinculada a la anulación
		if sriParams != nil {
			accessKey, err = d.sriService.EmitirNotaCredito(ctx, voidTxID, sriParams.OriginalTxID, sriParams.Motivo, sriParams.Password)
			if err != nil {
				// ERROR EN SRI: Revertir la anulación local para mantener consistencia
				d.logger.Printf("Fallo SRI (%v). Revirtiendo anulación local %d...", err, voidTxID)
				rollbackErr := d.service.RevertVoidTransaction(ctx, voidTxID)
				if rollbackErr != nil {
					// Peor escenario: Falló SRI y falló rollback. Inconsistencia manual requerida.
					d.logger.Printf("CRITICAL: Falló reversión de anulación %d: %v", voidTxID, rollbackErr)
					return fmt.Errorf("error SRI: %v. ADVERTENCIA: Falló la reversión local: %v", err, rollbackErr)
				}
				return fmt.Errorf("error SRI: %v. La anulación ha sido cancelada", err)
			}
		}
		return nil
	}

	// Definir callback de éxito (se ejecuta DESPUÉS de que el loading se cierra)
	onSuccess := func() {
		successMsg := "Transacción anulada con éxito."
		if sriParams != nil {
			successMsg = fmt.Sprintf("✅ Nota de Crédito Emitida.\n\nClave: %s\n\nEl comprobante está siendo procesado por el SRI y se enviará al cliente automáticamente al ser autorizado.", accessKey)
		}
		dialog.ShowInformation("Éxito", successMsg, d.mainWin)

		if d.callbackAction != nil {
			d.callbackAction()
		}
	}

	componets.HandleLongRunningOperation(d.mainWin, msg, operation, onSuccess)
}
