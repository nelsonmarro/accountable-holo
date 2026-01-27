package componets

import (
	"context"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
)

type EmissionPointService interface {
	GetEmissionPoints(ctx context.Context) ([]domain.EmissionPoint, error)
	UpdateEmissionPoint(ctx context.Context, ep *domain.EmissionPoint) error
}

type EmissionPointDialog struct {
	window  fyne.Window
	service EmissionPointService
	points  []domain.EmissionPoint
	list    *widget.List
}

func NewEmissionPointDialog(window fyne.Window, service EmissionPointService) *EmissionPointDialog {
	return &EmissionPointDialog{
		window:  window,
		service: service,
	}
}

func (d *EmissionPointDialog) Show() {
	// Intentar cargar datos antes de mostrar nada
	pts, err := d.service.GetEmissionPoints(context.Background())
	if err != nil || len(pts) == 0 {
		dialog.ShowInformation("Configuración Requerida", 
			"Para gestionar secuenciales, primero debe guardar los 'Datos Legales' y 'Códigos de Emisión' en la pestaña anterior.\n\nUna vez guardados, el sistema habilitará sus puntos de emisión automáticamente.", d.window)
		return
	}
	d.points = pts

	d.list = widget.NewList(
		func() int {
			return len(d.points)
		},
		d.makePointListUI,
		d.fillPointListData,
	)

	// Encabezado de Tabla
	header := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Tipo Doc.", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Est-Punto", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actual", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Inicial", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Acciones", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Gestión de Secuenciales y Migración", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			header,
		),
		nil, nil, nil,
		d.list,
	)

	customDialog := dialog.NewCustom("Puntos de Emisión SRI", "Cerrar", container.NewPadded(content), d.window)
	customDialog.Resize(fyne.NewSize(800, 500))
	customDialog.Show()
}

func (d *EmissionPointDialog) makePointListUI() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editBtn.Importance = widget.HighImportance

	return container.NewGridWithColumns(5,
		widget.NewLabel("Tipo"),
		widget.NewLabel("001-001"),
		widget.NewLabel("0"),
		widget.NewLabel("0"),
		container.NewCenter(editBtn),
	)
}

func (d *EmissionPointDialog) fillPointListData(i widget.ListItemID, o fyne.CanvasObject) {
	if i >= len(d.points) {
		return
	}
	p := d.points[i]
	row := o.(*fyne.Container)

	// Tipo
	name := "Factura (01)"
	if p.ReceiptType == "04" {
		name = "Nota de Crédito (04)"
	}
	row.Objects[0].(*widget.Label).SetText(name)

	// Codigo
	row.Objects[1].(*widget.Label).SetText(fmt.Sprintf("%s-%s", p.EstablishmentCode, p.EmissionPointCode))

	// Secuenciales
	row.Objects[2].(*widget.Label).SetText(fmt.Sprintf("%d", p.CurrentSequence))
	row.Objects[3].(*widget.Label).SetText(fmt.Sprintf("%d", p.InitialSequence))

	// Acciones
	actions := row.Objects[4].(*fyne.Container)
	editBtn := actions.Objects[0].(*widget.Button)
	editBtn.OnTapped = func() {
		d.editPoint(i)
	}
}

func (d *EmissionPointDialog) refreshData() {
	pts, err := d.service.GetEmissionPoints(context.Background())
	if err != nil {
		d.points = nil
		if d.list != nil {
			d.list.Refresh()
		}
		dialog.ShowInformation("Configuración Requerida", 
			"Para gestionar secuenciales, primero debe guardar los 'Datos Legales' y 'Códigos de Emisión' en la pestaña anterior.", d.window)
		return
	}
	d.points = pts
	if d.list != nil {
		d.list.Refresh()
	}
}

func (d *EmissionPointDialog) editPoint(row int) {
	p := d.points[row]

	currentEntry := widget.NewEntry()
	currentEntry.SetText(fmt.Sprintf("%d", p.CurrentSequence))

	initialEntry := widget.NewEntry()
	initialEntry.SetText(fmt.Sprintf("%d", p.InitialSequence))

	// Textos explicativos detallados con ejemplos
	helpCurrent := widget.NewLabelWithStyle("Representa el ÚLTIMO número de factura generado con éxito (en cualquier sistema).\nEl sistema sumará +1 a este valor para generar su siguiente comprobante.\n\nEjemplo: Si su última factura fue la 1500, ingrese 1500 aquí para que la próxima sea la 1501.", 
		fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	
	helpInitial := widget.NewLabelWithStyle("Representa el PRIMER número de factura emitido en Verith tras la migración.\nEste valor queda guardado como una marca histórica para auditorías futuras.\n\nEjemplo: Si su primera factura en Verith será la 1501, ingrese 1501 aquí.", 
		fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Configuración de Secuenciales", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		
		widget.NewLabel("Secuencial Actual (Contador de hoy):"),
		currentEntry,
		container.NewPadded(helpCurrent),
		
		widget.NewSeparator(),
		
		widget.NewLabel("Secuencial Inicial (Punto de Partida):"),
		initialEntry,
		container.NewPadded(helpInitial),
	)

	editDlg := dialog.NewCustomConfirm("Ajustar Secuenciales", "Guardar", "Cancelar", container.NewPadded(form), func(ok bool) {
		if ok {
			cur, err1 := strconv.Atoi(currentEntry.Text)
			ini, err2 := strconv.Atoi(initialEntry.Text)

			if err1 != nil || err2 != nil {
				dialog.ShowError(fmt.Errorf("los valores deben ser numéricos"), d.window)
				return
			}

			// Safety Check
			if cur != p.CurrentSequence {
				dialog.ShowConfirm("⚠ Advertencia Crítica", 
					fmt.Sprintf("Está a punto de cambiar el secuencial de facturación de %d a %d.\n\nSi reduce este número, podría generar facturas duplicadas que el SRI rechazará.\nSolo haga esto si está seguro de la sincronización con su sistema anterior.\n\n¿Desea continuar?", p.CurrentSequence, cur), 
					func(confirm bool) {
						if confirm {
							d.performUpdate(p, cur, ini)
						}
					}, d.window)
			} else {
				d.performUpdate(p, cur, ini)
			}
		}
	}, d.window)

	editDlg.Resize(fyne.NewSize(550, 500)) // Más grande para legibilidad
	editDlg.Show()
}
func (d *EmissionPointDialog) performUpdate(p domain.EmissionPoint, cur, ini int) {
	p.CurrentSequence = cur
	p.InitialSequence = ini

	err := d.service.UpdateEmissionPoint(context.Background(), &p)
	if err != nil {
		dialog.ShowError(err, d.window)
	} else {
		d.refreshData()
		dialog.ShowInformation("Actualizado", "Secuencial actualizado correctamente.", d.window)
	}
}