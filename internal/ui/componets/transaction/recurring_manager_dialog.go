package transaction

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type RecurringManagerDialog struct {
	window          fyne.Window
	service         RecurringTransactionService
	accountService  AccountService
	categoryService CategoryService
	data            []domain.RecurringTransaction
	table           *widget.Table
	selectedRow     int
}

func NewRecurringManagerDialog(
	parent fyne.Window,
	service RecurringTransactionService,
	accountService AccountService,
	categoryService CategoryService,
) *RecurringManagerDialog {
	d := &RecurringManagerDialog{
		window:          parent,
		service:         service,
		accountService:  accountService,
		categoryService: categoryService,
		selectedRow:     -1,
	}
	return d
}

func (d *RecurringManagerDialog) Show() {
	d.refreshData()

	d.table = widget.NewTable(
		func() (int, int) {
			return len(d.data), 5 // Description, Amount, Interval, NextRun, Status
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			if i.Row >= len(d.data) {
				return
			}
			item := d.data[i.Row]
			label := o.(*widget.Label)
			switch i.Col {
			case 0:
				label.SetText(item.Description)
			case 1:
				label.SetText(fmt.Sprintf("%.2f", item.Amount))
			case 2:
				label.SetText(string(item.Interval))
			case 3:
				label.SetText(item.NextRunDate.Format("02/01/2006"))
			case 4:
				status := "Activo"
				if !item.IsActive {
					status = "Pausado"
				}
				label.SetText(status)
			}
		},
	)

	d.table.OnSelected = func(id widget.TableCellID) {
		d.selectedRow = id.Row
	}
	d.table.OnUnselected = func(id widget.TableCellID) {
		if d.selectedRow == id.Row {
			d.selectedRow = -1
		}
	}

	// Set column widths
	d.table.SetColumnWidth(0, 250)
	d.table.SetColumnWidth(1, 100)
	d.table.SetColumnWidth(2, 100)
	d.table.SetColumnWidth(3, 100)
	d.table.SetColumnWidth(4, 80)

	// Toolbar
	toolbar := container.NewHBox(
		widget.NewButtonWithIcon("Nueva", theme.ContentAddIcon(), d.onNew),
		widget.NewButtonWithIcon("Editar", theme.DocumentCreateIcon(), d.onEdit),
		widget.NewButtonWithIcon("Pausar/Activar", theme.MediaPlayIcon(), d.onToggle),
		widget.NewButtonWithIcon("Eliminar", theme.DeleteIcon(), d.onDelete),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("Actualizar", theme.ViewRefreshIcon(), d.refreshData),
	)

	content := container.NewBorder(toolbar, nil, nil, nil, d.table)
	
	// Wrap in a container to give it size
	mainContainer := container.NewPadded(content)
	
	modal := dialog.NewCustom("Gestor de Transacciones Recurrentes", "Cerrar", mainContainer, d.window)
	modal.Resize(fyne.NewSize(800, 600))
	modal.Show()
}

func (d *RecurringManagerDialog) refreshData() {
	data, err := d.service.GetAll(context.Background())
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}
	d.data = data
	d.selectedRow = -1 // Reset selection on refresh
	if d.table != nil {
		d.table.Refresh()
		d.table.UnselectAll()
	}
}

func (d *RecurringManagerDialog) onNew() {
	form := NewRecurringForm(d.window, nil, d.service, d.accountService, d.categoryService)
	form.OnSaved = func() {
		d.refreshData()
	}
	form.Show()
}

func (d *RecurringManagerDialog) onEdit() {
	if d.selectedRow < 0 || d.selectedRow >= len(d.data) {
		return
	}
	item := d.data[d.selectedRow]
	form := NewRecurringForm(d.window, &item, d.service, d.accountService, d.categoryService)
	form.OnSaved = func() {
		d.refreshData()
	}
	form.Show()
}

func (d *RecurringManagerDialog) onToggle() {
	if d.selectedRow < 0 || d.selectedRow >= len(d.data) {
		return
	}
	item := d.data[d.selectedRow]
	item.IsActive = !item.IsActive
	
	err := d.service.Update(context.Background(), &item)
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}
	d.refreshData()
}

func (d *RecurringManagerDialog) onDelete() {
	if d.selectedRow < 0 || d.selectedRow >= len(d.data) {
		return
	}
	item := d.data[d.selectedRow]

	dialog.ShowConfirm("Confirmar", "¿Está seguro de eliminar esta regla recurrente?", func(ok bool) {
		if ok {
			err := d.service.Delete(context.Background(), item.ID)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
			d.refreshData()
		}
	}, d.window)
}
