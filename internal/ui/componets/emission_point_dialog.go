package componets

import (
	"context"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	list    *widget.Table
}

func NewEmissionPointDialog(window fyne.Window, service EmissionPointService) *EmissionPointDialog {
	return &EmissionPointDialog{
		window:  window,
		service: service,
	}
}

func (d *EmissionPointDialog) Show() {
	d.refreshData()

	d.list = widget.NewTable(
		func() (int, int) {
			return len(d.points), 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell Content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row >= len(d.points) {
				return
			}
			p := d.points[i.Row]
			l := o.(*widget.Label)
			switch i.Col {
			case 0:
				name := "Factura (01)"
				if p.ReceiptType == "04" {
					name = "Nota de Crédito (04)"
				}
				l.SetText(name)
			case 1:
				l.SetText(fmt.Sprintf("%s-%s", p.EstablishmentCode, p.EmissionPointCode))
			case 2:
				l.SetText(fmt.Sprintf("%d", p.CurrentSequence))
			case 3:
				l.SetText(fmt.Sprintf("%d", p.InitialSequence))
			}
		},
	)

	d.list.SetColumnWidth(0, 180)
	d.list.SetColumnWidth(1, 100)
	d.list.SetColumnWidth(2, 120)
	d.list.SetColumnWidth(3, 120)

	d.list.OnSelected = func(id widget.TableCellID) {
		d.editPoint(id.Row)
		d.list.UnselectAll()
	}

	content := container.NewBorder(
		widget.NewLabelWithStyle("Gestión de Secuenciales (Haga clic para editar)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		d.list,
	)

	customDialog := dialog.NewCustom("Puntos de Emisión SRI", "Cerrar", content, d.window)
	customDialog.Resize(fyne.NewSize(600, 400))
	customDialog.Show()
}

func (d *EmissionPointDialog) refreshData() {
	pts, err := d.service.GetEmissionPoints(context.Background())
	if err != nil {
		dialog.ShowError(err, d.window)
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

	items := []*widget.FormItem{
		{Text: "Secuencial Actual", Widget: currentEntry},
		{Text: "Secuencial Inicial (Migración)", Widget: initialEntry},
	}

	dialog.ShowForm("Editar Secuencial", "Guardar", "Cancelar", items, func(ok bool) {
		if ok {
			cur, _ := strconv.Atoi(currentEntry.Text)
			ini, _ := strconv.Atoi(initialEntry.Text)

			p.CurrentSequence = cur
			p.InitialSequence = ini

			err := d.service.UpdateEmissionPoint(context.Background(), &p)
			if err != nil {
				dialog.ShowError(err, d.window)
			} else {
				d.refreshData()
			}
		}
	}, d.window)
}
